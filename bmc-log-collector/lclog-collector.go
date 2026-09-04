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

// collectLifecycleLog collects the LC (Lifecycle) log from iDRAC.
//
// Unlike the SEL endpoint, the LC log endpoint returns only the latest page
// (50 entries on the real iDRAC, newest first). This function follows
// Members@odata.nextLink backward until it finds the entry read in the
// previous cycle. On the first collection for a machine, and after the LC
// log was cleared in iDRAC, only the latest page is emitted so that the
// whole history is not backfilled at once.
func (c *logCollector) collectLifecycleLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	filePath := path.Join(c.ptrDir, m.Serial)

	err := checkAndCreatePointerFile(filePath)
	if err != nil {
		slog.Error("can't check a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}

	lastPtr, err := readLastPointer(filePath)
	if err != nil {
		slog.Error("can't read a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}

	var page0 []LifeCycleLog // the latest page, kept for the first collection and the after-clear restart
	var newLogs []LifeCycleLog
	seen := make(map[string]struct{})
	newestId := 0
	var newestCreateTime int64
	foundKnown := false
	cleared := false
	exhausted := false // reached the end of the LC log without finding the last read entry

	bmcUrl := "https://" + m.BmcIP + c.rfLcPath
	page := 0
scan:
	for ; page < c.lcMaxPages; page++ {
		byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, bmcUrl)
		if err != nil {
			counterRequestFailed.WithLabelValues(m.Serial, metricLogTypeLc).Inc()
			// Prevent log output by the same error code
			if lastPtr.LcLastError != err.Error() {
				slog.Error("failed access to iDRAC on TCP/IP level.", "url", bmcUrl, "err", err.Error(), "serial", m.Serial)
			}
			lastPtr.LcLastHttpStatusCode = 0
			lastPtr.LcLastError = err.Error()
			if err := updateLastPointer(lastPtr, filePath); err != nil {
				slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
			}
			return
		}
		if page == 0 && (statusCode == http.StatusNotFound || statusCode == http.StatusMethodNotAllowed) {
			// The BMC does not implement the LC log service.
			// Not counted as a failure to avoid a permanent false alarm.
			if statusCode != lastPtr.LcLastHttpStatusCode {
				slog.Warn("the lifecycle log service is not implemented on this BMC", "url", bmcUrl, "httpStatusCode", statusCode, "serial", m.Serial)
			}
			lastPtr.LcLastHttpStatusCode = statusCode
			lastPtr.LcLastError = ""
			if err := updateLastPointer(lastPtr, filePath); err != nil {
				slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
			}
			return
		}
		if statusCode != http.StatusOK {
			counterRequestFailed.WithLabelValues(m.Serial, metricLogTypeLc).Inc()
			// Prevent log output by the same httpStatus
			if statusCode != lastPtr.LcLastHttpStatusCode {
				slog.Error("failed access to iDRAC on HTTP level.", "url", bmcUrl, "httpStatusCode", statusCode, "serial", m.Serial)
			}
			lastPtr.LcLastHttpStatusCode = statusCode
			lastPtr.LcLastError = ""
			if err := updateLastPointer(lastPtr, filePath); err != nil {
				slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
			}
			return
		}
		counterRequestSuccess.WithLabelValues(m.Serial, metricLogTypeLc).Inc()
		lastPtr.LcLastHttpStatusCode = statusCode
		lastPtr.LcLastError = ""

		var response RedfishLcLogSchema
		if err := json.Unmarshal(byteJSON, &response); err != nil {
			slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
			return
		}
		if len(response.Members) == 0 {
			exhausted = true
			break
		}

		if page == 0 {
			page0 = response.Members
			newestId, err = strconv.Atoi(response.Members[0].Id)
			if err != nil {
				slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", response.Members[0].Id)
				return
			}
			createTime, err := time.Parse(time.RFC3339, response.Members[0].Create)
			if err != nil {
				slog.Error("failed to parse for time; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Created", response.Members[0].Create)
				return
			}
			newestCreateTime = createTime.Unix()

			// The first collection for this machine: start from the latest page without backfill
			if lastPtr.LcLastReadId == 0 {
				break
			}
			// The entry Id restarts from 1 when the LC log is cleared in iDRAC
			if newestId < lastPtr.LcLastReadId {
				cleared = true
				break
			}
		}

		for _, v := range response.Members {
			// The Id and the Created time are the basis of the pointer management.
			// When they cannot be parsed, abort this cycle without advancing the
			// pointer so that no entry is skipped permanently; the next cycle retries.
			id, err := strconv.Atoi(v.Id)
			if err != nil {
				slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", v.Id)
				return
			}
			if id == lastPtr.LcLastReadId && lastPtr.LcLastReadCreateTime != 0 {
				createTime, err := time.Parse(time.RFC3339, v.Create)
				if err != nil {
					slog.Error("failed to parse for time; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Created", v.Create)
					return
				}
				if createTime.Unix() != lastPtr.LcLastReadCreateTime {
					// The same Id with a different creation time: the LC log was
					// cleared and has grown beyond the last read Id since then
					cleared = true
					break scan
				}
			}
			if id <= lastPtr.LcLastReadId {
				foundKnown = true
				break scan
			}
			// An entry created between the page requests shifts the pages backward
			// and the next page repeats the entries of the previous page.
			// Skip the already collected entries.
			if _, ok := seen[v.Id]; ok {
				continue
			}
			seen[v.Id] = struct{}{}
			newLogs = append(newLogs, v)
		}

		if response.NextLink == "" {
			// The last page of the LC log
			exhausted = true
			break
		}
		bmcUrl = "https://" + m.BmcIP + response.NextLink
	}

	var emitErr error
	switch {
	case lastPtr.LcLastReadId == 0:
		emitErr = c.emitLifecycleLogs(page0, m, logWriter)
	case cleared:
		slog.Warn("the lifecycle log was cleared in iDRAC; restarting from the latest page", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "newestId", newestId)
		emitErr = c.emitLifecycleLogs(page0, m, logWriter)
	default:
		if !foundKnown && len(newLogs) > 0 {
			if page == c.lcMaxPages {
				counterLcCatchupTruncated.WithLabelValues(m.Serial).Inc()
				slog.Warn("stopped catching up the lifecycle log at the page limit; the entries in between are skipped", "serial", m.Serial, "pageLimit", c.lcMaxPages, "lastReadId", lastPtr.LcLastReadId, "newestId", newestId)
			} else if exhausted {
				slog.Warn("reached the end of the lifecycle log without finding the last read entry; the log may have been cleared", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "newestId", newestId)
			}
		}
		emitErr = c.emitLifecycleLogs(newLogs, m, logWriter)
	}

	// Advance the pointer only when all the entries were written, so that a
	// write failure does not lose entries; the next cycle re-emits them
	if emitErr == nil && newestId > 0 {
		lastPtr.LcLastReadId = newestId
		lastPtr.LcLastReadCreateTime = newestCreateTime
	}
	err = updateLastPointer(lastPtr, filePath)
	if err != nil {
		slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
	}
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
