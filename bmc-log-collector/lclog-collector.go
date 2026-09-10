package main

import (
	"cmp"
	"context"
	"encoding/json"
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

// RedfishLcLogSchema is the Lifecycle log entry collection as returned by
// Redfish: the latest page of the log.
type RedfishLcLogSchema struct {
	Name        string         `json:"Name"`
	Count       int            `json:"Members@odata.count"`
	Context     string         `json:"@odata.context"`
	Id          string         `json:"@odata.id"`
	Type        string         `json:"@odata.type"`
	Description string         `json:"Description"`
	Members     []LifeCycleLog `json:"Members"`
}

// lcEntry is a Lifecycle log entry with its Id parsed as a number.
type lcEntry struct {
	id int
	LifeCycleLog
}

// collectLifecycleLog collects the LC (Lifecycle) log from iDRAC in the same
// way as the SEL: the entries newer than the Id recorded in the pointer file
// are emitted, and the pointer is advanced to the newest Id.
//
// Unlike the SEL endpoint, the LC log endpoint returns only the latest page
// (50 entries on the real iDRAC). The entries that fell off the page since the
// previous cycle are not collected; this is accepted because the LC log grows
// only a few entries per day in our fleet (see docs/design.md).
func (c *logCollector) collectLifecycleLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	filePath := path.Join(c.ptrDir, m.Serial)

	lastPtr, err := loadLastPointer(filePath)
	if err != nil {
		slog.Error("can't load a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}

	bmcUrl := "https://" + m.BmcIP + c.rfLcPath
	// A 404/405 reply means that the BMC does not implement the LC log service
	byteJSON, err := c.requestBmcLog(ctx, m, bmcUrl, metricLogTypeLc, &lastPtr.LcLastHttpStatusCode, &lastPtr.LcLastError, http.StatusNotFound, http.StatusMethodNotAllowed)
	if err != nil {
		// The failure has been reported; record the request status and keep
		// the read position unchanged so that the next cycle retries
		if err := updateLastPointer(lastPtr, filePath); err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		}
		return
	}
	// Clear the failure status so that the same failure after a recovery is
	// reported again instead of being suppressed by the deduplication
	lastPtr.LcLastHttpStatusCode = http.StatusOK
	lastPtr.LcLastError = ""

	var response RedfishLcLogSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	// The Id is the basis of the pointer management. Validate all the Ids
	// before writing any entry: aborting after some entries were written
	// would re-emit them every cycle while a malformed entry persists.
	entries := make([]lcEntry, len(response.Members))
	for i, v := range response.Members {
		id, err := strconv.Atoi(v.Id)
		if err != nil {
			slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", v.Id, "ptrDir", c.ptrDir)
			return
		}
		entries[i] = lcEntry{id: id, LifeCycleLog: v}
	}
	if len(entries) == 0 {
		// The LC log is empty; there is nothing to collect. A clear is
		// detected by the Id when new entries arrive.
		if err := updateLastPointer(lastPtr, filePath); err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		}
		return
	}
	// Emit in the ascending order of the Id. The real iDRAC returns the
	// entries newest first, but the order is not relied on.
	slices.SortFunc(entries, func(a, b lcEntry) int { return cmp.Compare(a.id, b.id) })
	oldest, newest := entries[0], entries[len(entries)-1]

	newestCreateTime, err := time.Parse(time.RFC3339, newest.Create)
	if err != nil {
		slog.Error("failed to parse for time", "err", err, "serial", m.Serial, "Id", newest.Id)
		return
	}

	cleared, err := isLcLogCleared(lastPtr, entries)
	if err != nil {
		slog.Error("failed to parse for time; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial)
		return
	}
	if cleared {
		slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "newestId", newest.id)
	} else if lastPtr.LcLastReadId > 0 && oldest.id > lastPtr.LcLastReadId+1 {
		// The entries between the pointer and the page fell off the page
		slog.Warn("the entries between the last read entry and the latest page were not collected", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "oldestId", oldest.id, "newestId", newest.id)
	}

	for _, e := range entries {
		// Output duplicate log, after log clear in iDRAC
		if e.id <= lastPtr.LcLastReadId && !cleared {
			continue
		}
		v := e.LifeCycleLog
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
			// Abort without updating the pointer file so that the entry is
			// not lost; the next cycle re-emits from the last persisted Id
			slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", e.id, "ptrDir", c.ptrDir)
			return
		}

		lastPtr.LcLastReadId = e.id
	}

	// All the entries up to the newest one were written
	lastPtr.LcLastReadCreateTime = newestCreateTime.Unix()
	if err := updateLastPointer(lastPtr, filePath); err != nil {
		slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
	}
}

// isLcLogCleared reports whether the LC log was cleared in iDRAC since the
// previous cycle. The entry Id restarts from 1 on a clear, so the log was
// cleared when the newest Id is smaller than the pointer, or when the entry
// with the pointered Id has a different creation time (the log was cleared
// and has grown beyond the pointer since then). entries must be sorted in
// the ascending order of the Id and not empty.
func isLcLogCleared(lastPtr LastPointer, entries []lcEntry) (bool, error) {
	if lastPtr.LcLastReadId == 0 {
		// The first collection for the machine
		return false, nil
	}
	if entries[len(entries)-1].id < lastPtr.LcLastReadId {
		return true, nil
	}
	if lastPtr.LcLastReadCreateTime == 0 {
		// The pointer file was written by an older version
		return false, nil
	}
	for _, e := range entries {
		if e.id != lastPtr.LcLastReadId {
			continue
		}
		createTime, err := time.Parse(time.RFC3339, e.Create)
		if err != nil {
			return false, fmt.Errorf("parse Created %q of Id %s: %w", e.Create, e.Id, err)
		}
		return createTime.Unix() != lastPtr.LcLastReadCreateTime, nil
	}
	return false, nil
}
