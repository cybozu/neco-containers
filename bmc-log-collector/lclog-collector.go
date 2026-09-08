package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"time"
)

// LifeCycleLog is an entry of the iDRAC Lifecycle log as returned by Redfish,
// extended with the fields that identify the machine in the output.
type LifeCycleLog struct {
	ODataID          string       `json:"@odata.id"`
	ODataType        string       `json:"@odata.type"`
	Create           string       `json:"Created"`
	Description      string       `json:"Description"`
	EntryType        string       `json:"EntryType"`
	Id               string       `json:"Id"`
	Message          string       `json:"Message"`
	MessageArgs      []string     `json:"MessageArgs"`
	MessageArgsCount int          `json:"MessageArgs@odata.count"`
	MessageId        string       `json:"MessageId"`
	Name             string       `json:"Name"`
	Oem              LifeCycleOem `json:"Oem"`
	OemRecordFormat  string       `json:"OemRecordFormat"`
	Severity         string       `json:"Severity"`
	Serial           string
	NodeIP           string
	BmcIP            string
	LogType          string
}

// LifeCycleOem is the vendor-specific part of a Lifecycle log entry.
type LifeCycleOem struct {
	Dell LifeCycleOemDell `json:"Dell"`
}

// LifeCycleOemDell is the Dell-specific part of a Lifecycle log entry.
type LifeCycleOemDell struct {
	ODataType         string  `json:"@odata.type"`
	Category          string  `json:"Category"`
	Comment           *string `json:"Comment"`
	LastUpdatedByUser *string `json:"LastUpdatedByUser"`
}

// RedfishLcLogSchema is one page of the Lifecycle log entry collection.
type RedfishLcLogSchema struct {
	Name        string         `json:"Name"`
	Count       int            `json:"Members@odata.count"`
	Context     string         `json:"@odata.context"`
	Id          string         `json:"@odata.id"`
	Type        string         `json:"@odata.type"`
	Description string         `json:"Description"`
	Members     []LifeCycleLog `json:"Members"`
	NextLink    string         `json:"Members@odata.nextLink"`
}

// lcScanResult is the outcome of a successful scanLifecycleLog.
type lcScanResult struct {
	logs             []LifeCycleLog // the entries to emit, newest first
	newestId         int            // Id of the newest entry; 0 when the log is empty
	newestCreateTime int64          // Created time of the newest entry
	gap              bool           // the scan stopped at the page limit before reaching the target: the entries in between are not in logs
}

// lcScanTarget is the entry read in the previous cycle, where the scan stops.
type lcScanTarget struct {
	id         int
	createTime int64 // 0 when unknown (pointer file written by an older version)
}

// collectLifecycleLog collects the LC (Lifecycle) log from iDRAC.
//
// Unlike the SEL endpoint, the LC log endpoint returns only the latest page
// (50 entries on the real iDRAC, newest first). This function follows
// Members@odata.nextLink backward until it finds the entry read in the
// previous cycle, up to lcMaxPages pages. The first collection for a machine
// and the collection after the LC log was cleared in iDRAC read only the
// latest page, so that the whole history is not ingested at once.
func (c *logCollector) collectLifecycleLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	filePath := path.Join(c.ptrDir, m.Serial)

	lastPtr, err := loadLastPointer(filePath)
	if err != nil {
		slog.Error("can't load a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}
	lastReadId := lastPtr.LcLastReadId

	result, cycleErr := c.updateLifecycleLog(ctx, m, &lastPtr, logWriter)
	if cycleErr != nil && !errors.Is(cycleErr, errBMCRequestFailed) {
		slog.Error("failed to collect the lifecycle log; the read position is kept for a retry", "err", cycleErr, "serial", m.Serial)
	}
	// The pointer file is written on every outcome: the request status is
	// recorded even when the cycle failed, while the read position has been
	// advanced only after all the new entries were written
	if err := updateLastPointer(lastPtr, filePath); err != nil {
		slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}
	// The gap is reported only now that the new position is persisted; a
	// failure above makes the next cycle retry from the old position instead
	if cycleErr == nil && result.gap {
		counterLcPageLimitReached.WithLabelValues(m.Serial, metricLogTypeLc).Inc()
		slog.Warn("stopped catching up the lifecycle log at the page limit; the entries in between are skipped", "serial", m.Serial, "pageLimit", c.lcMaxPages, "lastReadId", lastReadId, "newestId", result.newestId)
	}
}

// updateLifecycleLog scans the new entries, writes them, and advances the
// read position in lastPtr. On an error the read position is left unchanged
// so that the next cycle retries; errBMCRequestFailed means that the
// failure has already been reported by requestBmcLog.
func (c *logCollector) updateLifecycleLog(ctx context.Context, m Machine, lastPtr *LastPointer, logWriter bmcLogWriter) (lcScanResult, error) {
	result, err := c.scanLifecycleLog(ctx, m, lastPtr)
	if err != nil {
		return lcScanResult{}, err
	}
	// The whole scan succeeded: clear the failure status so that the same
	// failure after a recovery is reported again
	lastPtr.LcLastHttpStatusCode = http.StatusOK
	lastPtr.LcLastError = ""

	if err := c.emitLifecycleLogs(result.logs, m, logWriter); err != nil {
		return lcScanResult{}, err
	}
	if result.newestId > 0 {
		lastPtr.LcLastReadId = result.newestId
		lastPtr.LcLastReadCreateTime = result.newestCreateTime
	}
	return result, nil
}

// scanLifecycleLog follows the LC log pages from the newest entry backward
// and gathers the entries newer than the target, the entry read in the
// previous cycle. Without a target (the first collection for a machine) or
// when the LC log was cleared in iDRAC since the previous cycle, it returns
// only the latest page.
//
// The Id and the Created time are the basis of the pointer management, so a
// page whose Ids cannot be parsed or are not in the newest-first order, or an
// entry whose Created time cannot be parsed, is an error: the cycle is
// aborted to avoid skipping entries permanently. lastPtr carries the status
// of a failed request.
func (c *logCollector) scanLifecycleLog(ctx context.Context, m Machine, lastPtr *LastPointer) (lcScanResult, error) {
	var target *lcScanTarget
	if lastPtr.LcLastReadId > 0 {
		target = &lcScanTarget{id: lastPtr.LcLastReadId, createTime: lastPtr.LcLastReadCreateTime}
	}

	var result lcScanResult
	var latestPage []LifeCycleLog
	seen := make(map[string]struct{})
	url := "https://" + m.BmcIP + c.rfLcPath
	for page := 0; page < c.lcMaxPages; page++ {
		response, err := c.fetchLifecycleLogPage(ctx, m, lastPtr, url, page == 0)
		if err != nil {
			return lcScanResult{}, fmt.Errorf("page %d: %w", page, err)
		}
		ids, err := lcPageIds(response.Members)
		if err != nil {
			return lcScanResult{}, fmt.Errorf("page %d: %w", page, err)
		}

		if page == 0 {
			if len(response.Members) == 0 {
				// The LC log is empty
				return lcScanResult{}, nil
			}
			latestPage = response.Members
			newestCreateTime, err := parseLifecycleLogCreateTime(latestPage[0])
			if err != nil {
				return lcScanResult{}, err
			}
			result.newestId = ids[0]
			result.newestCreateTime = newestCreateTime.Unix()

			if target == nil {
				// The first collection for the machine: only the latest page
				result.logs = latestPage
				return result, nil
			}
			// The entry Id restarts from 1 when the LC log is cleared in iDRAC
			if ids[0] < target.id {
				slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", target.id, "newestId", ids[0])
				result.logs = latestPage
				return result, nil
			}
		}

		for i, v := range response.Members {
			id := ids[i]
			if id == target.id && target.createTime != 0 {
				createTime, err := parseLifecycleLogCreateTime(v)
				if err != nil {
					return lcScanResult{}, err
				}
				if createTime.Unix() != target.createTime {
					// The same Id with a different creation time: the LC log was
					// cleared and has grown beyond the target since then. Only the
					// latest page is collected, as on the first collection.
					slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", target.id, "Id", v.Id)
					result.logs = latestPage
					return result, nil
				}
			}
			if id <= target.id {
				// Reached the target: the entries gathered so far are the new ones
				return result, nil
			}
			// An entry created between the page requests shifts the pages backward
			// and the next page repeats the entries of the previous page.
			// Skip the already collected entries.
			if _, dup := seen[v.Id]; dup {
				continue
			}
			seen[v.Id] = struct{}{}
			result.logs = append(result.logs, v)
		}

		if response.NextLink == "" || len(response.Members) == 0 {
			// The last page of the LC log
			slog.Warn("reached the end of the lifecycle log without finding the last read entry; the log may have been cleared", "serial", m.Serial, "lastReadId", target.id, "newestId", result.newestId)
			return result, nil
		}
		url = "https://" + m.BmcIP + response.NextLink
	}

	// The target was not reached within the page limit
	result.gap = true
	return result, nil
}

// fetchLifecycleLogPage requests one page of the LC log. A request failure
// has been reported, counted, and recorded in lastPtr by requestBmcLog and is
// returned as errBMCRequestFailed. A 404/405 reply to the first page means
// that the BMC does not implement the LC log service; the same reply to a
// later page is an ordinary failure.
func (c *logCollector) fetchLifecycleLogPage(ctx context.Context, m Machine, lastPtr *LastPointer, url string, firstPage bool) (RedfishLcLogSchema, error) {
	var notImplemented []int
	if firstPage {
		notImplemented = []int{http.StatusNotFound, http.StatusMethodNotAllowed}
	}
	byteJSON, err := c.requestBmcLog(ctx, m, url, metricLogTypeLc, &lastPtr.LcLastHttpStatusCode, &lastPtr.LcLastError, notImplemented...)
	if err != nil {
		return RedfishLcLogSchema{}, err
	}

	var response RedfishLcLogSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		return RedfishLcLogSchema{}, fmt.Errorf("decode the response of %s: %w", url, err)
	}
	return response, nil
}

// lcPageIds parses the Ids of a page and verifies the newest-first order that
// the scan relies on: on the real iDRAC the Id is a number that decreases
// strictly along the page.
func lcPageIds(members []LifeCycleLog) ([]int, error) {
	ids := make([]int, len(members))
	for i, v := range members {
		id, err := strconv.Atoi(v.Id)
		if err != nil {
			return nil, fmt.Errorf("parse Id %q: %w", v.Id, err)
		}
		if i > 0 && id >= ids[i-1] {
			return nil, fmt.Errorf("the entries are not in the newest-first order: Id %d follows Id %d", id, ids[i-1])
		}
		ids[i] = id
	}
	return ids, nil
}

// parseLifecycleLogCreateTime parses the Created time of an LC log entry.
func parseLifecycleLogCreateTime(v LifeCycleLog) (time.Time, error) {
	createTime, err := time.Parse(time.RFC3339, v.Create)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse Created %q of Id %s: %w", v.Create, v.Id, err)
	}
	return createTime, nil
}

// emitLifecycleLogs writes the entries, given newest first, in ascending order.
// A write failure stops the emission and is returned so that the caller keeps
// the pointer unchanged; an entry that cannot be marshaled is skipped instead
// because retrying cannot fix it.
func (c *logCollector) emitLifecycleLogs(logs []LifeCycleLog, m Machine, logWriter bmcLogWriter) error {
	for _, v := range slices.Backward(logs) {
		// Add the information to identify of the node
		v.Serial = m.Serial
		v.BmcIP = m.BmcIP
		v.NodeIP = m.NodeIP
		v.LogType = logTypeLCLog

		bmcByteJsonLog, err := json.Marshal(v)
		if err != nil {
			slog.Error("failed to marshal the lifecycle log", "err", err, "serial", m.Serial, "Id", v.Id)
			continue
		}
		if err := logWriter.write(string(bmcByteJsonLog), m.Serial); err != nil {
			return fmt.Errorf("write the entry Id %s: %w", v.Id, err)
		}
	}
	return nil
}
