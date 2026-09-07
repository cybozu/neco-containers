package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"time"
)

type LifeCycleLog struct {
	Od_Id             string       `json:"@odata.id"`
	Od_Type           string       `json:"@odata.type"`
	Create            string       `json:"Created"`
	Description       string       `json:"Description"`
	EntryType         string       `json:"EntryType"`
	Id                string       `json:"Id"`
	Message           string       `json:"Message"`
	MessageArgs       []string     `json:"MessageArgs"`
	OdCnt_MessageArgs int          `json:"MessageArgs@odata.count"`
	MessageId         string       `json:"MessageId"`
	Name              string       `json:"Name"`
	Oem               LifeCycleOem `json:"Oem"`
	OemRecordFormat   string       `json:"OemRecordFormat"`
	Severity          string       `json:"Severity"`
	Serial            string
	NodeIP            string
	BmcIP             string
	LogType           string
}

type LifeCycleOem struct {
	Dell LifeCycleOemDell `json:"Dell"`
}

type LifeCycleOemDell struct {
	Od_Type           string  `json:"@odata.type"`
	Category          string  `json:"Category"`
	Comment           *string `json:"Comment"`
	LastUpdatedByUser *string `json:"LastUpdatedByUser"`
}

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

	result, ok := c.scanLifecycleLog(ctx, m, &lastPtr)
	if !ok {
		// The failure has been reported; record the request status and keep
		// the read position so that the next cycle retries
		saveLastPointer(lastPtr, filePath, m.Serial)
		return
	}
	// The whole scan succeeded: clear the failure status so that the same
	// failure after a recovery is reported again
	lastPtr.LcLastHttpStatusCode = http.StatusOK
	lastPtr.LcLastError = ""

	// Advance the read position only when all the entries were written, so
	// that a write failure does not lose entries; the next cycle re-emits them
	if err := c.emitLifecycleLogs(result.logs, m, logWriter); err != nil {
		saveLastPointer(lastPtr, filePath, m.Serial)
		return
	}
	if result.newestId > 0 {
		lastPtr.LcLastReadId = result.newestId
		lastPtr.LcLastReadCreateTime = result.newestCreateTime
	}
	if !saveLastPointer(lastPtr, filePath, m.Serial) {
		return
	}
	// The gap is reported only now that the new position is persisted; a
	// failure above makes the next cycle retry from the old position instead
	if result.gap {
		counterLcPageLimitReached.WithLabelValues(m.Serial).Inc()
		slog.Warn("stopped catching up the lifecycle log at the page limit; the entries in between are skipped", "serial", m.Serial, "pageLimit", c.lcMaxPages, "lastReadId", lastReadId, "newestId", result.newestId)
	}
}

// scanLifecycleLog follows the LC log pages from the newest entry backward
// and gathers the entries newer than the target, the entry read in the
// previous cycle. Without a target (the first collection for a machine) or
// when the LC log was cleared in iDRAC since the previous cycle, it returns
// only the latest page.
//
// It returns false when the cycle must be aborted: the failure has been
// reported and lastPtr carries the status of the failed request. The Id and
// the Created time are the basis of the pointer management, so an entry
// whose Id or Created time cannot be parsed aborts the cycle to avoid
// skipping it permanently.
func (c *logCollector) scanLifecycleLog(ctx context.Context, m Machine, lastPtr *LastPointer) (lcScanResult, bool) {
	var target *lcScanTarget
	if lastPtr.LcLastReadId > 0 {
		target = &lcScanTarget{id: lastPtr.LcLastReadId, createTime: lastPtr.LcLastReadCreateTime}
	}

	var result lcScanResult
	var latestPage []LifeCycleLog
	seen := make(map[string]struct{})
	url := "https://" + m.BmcIP + c.rfLcPath
	for page := 0; page < c.lcMaxPages; page++ {
		response, ok := c.fetchLifecycleLogPage(ctx, m, lastPtr, url, page == 0)
		if !ok {
			return lcScanResult{}, false
		}

		if page == 0 {
			if len(response.Members) == 0 {
				// The LC log is empty
				return lcScanResult{}, true
			}
			latestPage = response.Members
			newest := latestPage[0]
			newestId, ok := parseLifecycleLogId(m, newest)
			if !ok {
				return lcScanResult{}, false
			}
			newestCreateTime, ok := parseLifecycleLogCreateTime(m, newest)
			if !ok {
				return lcScanResult{}, false
			}
			result.newestId = newestId
			result.newestCreateTime = newestCreateTime.Unix()

			if target == nil {
				// The first collection for the machine
				return lcLatestPage(m, result, latestPage)
			}
			// The entry Id restarts from 1 when the LC log is cleared in iDRAC
			if newestId < target.id {
				slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", target.id, "newestId", newestId)
				return lcLatestPage(m, result, latestPage)
			}
		}

		for _, v := range response.Members {
			id, ok := parseLifecycleLogId(m, v)
			if !ok {
				return lcScanResult{}, false
			}
			if id == target.id && target.createTime != 0 {
				createTime, ok := parseLifecycleLogCreateTime(m, v)
				if !ok {
					return lcScanResult{}, false
				}
				if createTime.Unix() != target.createTime {
					// The same Id with a different creation time: the LC log was
					// cleared and has grown beyond the target since then
					slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", target.id, "Id", v.Id)
					return lcLatestPage(m, result, latestPage)
				}
			}
			if id <= target.id {
				// Reached the target: the entries gathered so far are the new ones
				return result, true
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
			return result, true
		}
		url = "https://" + m.BmcIP + response.NextLink
	}

	// The target was not reached within the page limit
	result.gap = true
	return result, true
}

// lcLatestPage returns the whole latest page as the entries to emit, for the
// cases without a previous position to catch up from. Every Id is validated
// first so that a malformed entry aborts the cycle instead of being skipped
// permanently when the position advances.
func lcLatestPage(m Machine, result lcScanResult, latestPage []LifeCycleLog) (lcScanResult, bool) {
	for _, v := range latestPage {
		if _, ok := parseLifecycleLogId(m, v); !ok {
			return lcScanResult{}, false
		}
	}
	result.logs = latestPage
	return result, true
}

// fetchLifecycleLogPage requests one page of the LC log. It returns false
// when the page could not be obtained; the failure has been reported and
// counted by requestBmcLog, and recorded in lastPtr.
// A 404/405 reply to the first page means that the BMC does not implement
// the LC log service; the same reply to a later page is an ordinary failure.
func (c *logCollector) fetchLifecycleLogPage(ctx context.Context, m Machine, lastPtr *LastPointer, url string, firstPage bool) (RedfishLcLogSchema, bool) {
	var notImplemented []int
	if firstPage {
		notImplemented = []int{http.StatusNotFound, http.StatusMethodNotAllowed}
	}
	byteJSON, ok := c.requestBmcLog(ctx, m, url, metricLogTypeLc, &lastPtr.LcLastHttpStatusCode, &lastPtr.LcLastError, notImplemented...)
	if !ok {
		return RedfishLcLogSchema{}, false
	}

	var response RedfishLcLogSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "url", url)
		return RedfishLcLogSchema{}, false
	}
	return response, true
}

// parseLifecycleLogId parses the numeric Id of an LC log entry. It returns
// false after reporting the failure.
func parseLifecycleLogId(m Machine, v LifeCycleLog) (int, bool) {
	id, err := strconv.Atoi(v.Id)
	if err != nil {
		slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", v.Id)
		return 0, false
	}
	return id, true
}

// parseLifecycleLogCreateTime parses the Created time of an LC log entry. It
// returns false after reporting the failure.
func parseLifecycleLogCreateTime(m Machine, v LifeCycleLog) (time.Time, bool) {
	createTime, err := time.Parse(time.RFC3339, v.Create)
	if err != nil {
		slog.Error("failed to parse for time; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Created", v.Create)
		return time.Time{}, false
	}
	return createTime, true
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
		err = logWriter.write(string(bmcByteJsonLog), m.Serial)
		if err != nil {
			slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog))
			return err
		}
	}
	return nil
}
