package main

import (
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

// LifecycleLog is an entry of the iDRAC Lifecycle log.
type LifecycleLog struct {
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
	Oem              LifecycleOem `json:"Oem"`
	OemRecordFormat  string       `json:"OemRecordFormat"`
	Severity         string       `json:"Severity"`
	Serial           string
	NodeIP           string
	BmcIP            string
	LogType          string
}

// LifecycleOem is the vendor-specific part of a Lifecycle log entry.
type LifecycleOem struct {
	Dell LifecycleOemDell `json:"Dell"`
}

// LifecycleOemDell is the Dell-specific part of a Lifecycle log entry.
type LifecycleOemDell struct {
	ODataType         string  `json:"@odata.type"`
	Category          string  `json:"Category"`
	Comment           *string `json:"Comment"`
	LastUpdatedByUser *string `json:"LastUpdatedByUser"`
}

// RedfishLcLogSchema is the latest page of the Lifecycle log entry collection.
type RedfishLcLogSchema struct {
	Name        string         `json:"Name"`
	Count       int            `json:"Members@odata.count"`
	Context     string         `json:"@odata.context"`
	Id          string         `json:"@odata.id"`
	Type        string         `json:"@odata.type"`
	Description string         `json:"Description"`
	Members     []LifecycleLog `json:"Members"`
}

// lcEntry is a Lifecycle log entry with its Id parsed as a number.
type lcEntry struct {
	idNum int
	LifecycleLog
}

// collectLifecycleLog collects the LC (Lifecycle) log in the same way as the SEL.
// The endpoint returns only the latest page (50 entries); the entries that fell
// off the page since the previous cycle are not collected (see docs/design.md).
func (c *logCollector) collectLifecycleLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	filePath := path.Join(c.ptrDir, m.Serial)

	lastPtr, err := loadLastPointer(filePath)
	if err != nil {
		slog.Error("can't load a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}

	bmcUrl := "https://" + m.BmcIP + c.rfLcPath
	byteJSON, err := c.requestBmcLog(ctx, m, bmcUrl, metricLogTypeLc, &lastPtr.LcLastHttpStatusCode, &lastPtr.LcLastError, http.StatusNotFound, http.StatusMethodNotAllowed)
	if err != nil {
		saveLastPointer(lastPtr, filePath, m.Serial)
		return
	}
	lastPtr.LcLastHttpStatusCode = http.StatusOK
	lastPtr.LcLastError = ""

	var response RedfishLcLogSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	// Unlike the SEL, an empty LC log is not an error: the log was just cleared
	if len(response.Members) == 0 {
		saveLastPointer(lastPtr, filePath, m.Serial)
		return
	}

	// Validate all the Ids before writing so that the cycle does not abort halfway
	entries := make([]lcEntry, len(response.Members))
	for i, v := range response.Members {
		id, err := strconv.Atoi(v.Id)
		if err != nil {
			slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", v.Id, "ptrDir", c.ptrDir)
			return
		}
		entries[i] = lcEntry{idNum: id, LifecycleLog: v}
	}
	newest, oldest := entries[0], entries[len(entries)-1]

	createTime, err := time.Parse(time.RFC3339, newest.Create)
	if err != nil {
		slog.Error("failed to parse for time", "err", err, "serial", m.Serial, "Id", newest.Id)
		return
	}
	newestCreateTime := createTime.Unix()

	cleared, err := isLcLogCleared(lastPtr, entries)
	if err != nil {
		slog.Error("failed to parse for time", "err", err, "serial", m.Serial)
		return
	}
	if cleared {
		slog.Warn("the lifecycle log was cleared in iDRAC; collecting the latest page", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "newestId", newest.idNum)
	} else if lastPtr.LcLastReadId > 0 && oldest.idNum > lastPtr.LcLastReadId+1 {
		slog.Warn("the entries between the last read entry and the latest page were not collected", "serial", m.Serial, "lastReadId", lastPtr.LcLastReadId, "oldestId", oldest.idNum, "newestId", newest.idNum)
	}

	for _, e := range slices.Backward(entries) {
		// Output duplicate log, after log clear in iDRAC
		if e.idNum <= lastPtr.LcLastReadId && !cleared {
			continue
		}
		v := e.LifecycleLog
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
			slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", e.idNum, "ptrDir", c.ptrDir)
			return
		}

		lastPtr.LcLastReadId = e.idNum
	}

	lastPtr.LcLastReadCreateTime = newestCreateTime
	saveLastPointer(lastPtr, filePath, m.Serial)
}

// isLcLogCleared reports whether the LC log was cleared in iDRAC, which
// restarts the Id from 1. The SEL compares the creation time of the oldest
// entry, but the oldest entry of the LC log page slides, so the creation time
// of the last read entry is compared instead. entries must be in the
// newest-first order and not empty.
func isLcLogCleared(lastPtr LastPointer, entries []lcEntry) (bool, error) {
	if entries[0].idNum < lastPtr.LcLastReadId {
		return true, nil
	}
	for _, e := range entries {
		if e.idNum != lastPtr.LcLastReadId {
			continue
		}
		createTime, err := time.Parse(time.RFC3339, e.Create)
		if err != nil {
			return false, fmt.Errorf("parse the creation time of the last read entry Id %s: %w", e.Id, err)
		}
		return createTime.Unix() != lastPtr.LcLastReadCreateTime, nil
	}
	return false, nil
}
