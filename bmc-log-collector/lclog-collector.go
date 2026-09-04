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

type lcScanResult struct {
	logs             []LifeCycleLog
	newestId         int
	newestCreateTime int64
}

type lcScanTarget struct {
	Id         int
	createTime int64
}

// collectLifecycleLog collects the LC (Lifecycle) log from iDRAC.
//
// Unlike the SEL endpoint, the LC log endpoint returns only the latest page
// (50 entries on the real iDRAC, newest first). This function follows
// Members@odata.nextLink backward until it finds the entry read in the
// previous cycle, up to lcMaxPages pages.
func (c *logCollector) collectLifecycleLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	filePath := path.Join(c.ptrDir, m.Serial)

	lastPtr, err := loadLastPointer(filePath)
	if err != nil {
		slog.Error("can't load a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		return
	}

	defer func() {
		if err := updateLastPointer(lastPtr, filePath); err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		}
	}()

	result, lastPtr, ok := c.scanLifecycleLog(ctx, m, lastPtr)
	if !ok {
		return
	}

	if err := c.emitLifecycleLogs(result.logs, m, logWriter); err != nil {
		return
	}

	if result.newestId > 0 {
		lastPtr.LcLastReadId = result.newestId
		lastPtr.LcLastReadCreateTime = result.newestCreateTime
	}
}

// scanLifecycleLog follows the LC log pages from the newest entry backward and
// gathers the entries newer than the one read in the previous cycle.
func (c *logCollector) scanLifecycleLog(ctx context.Context, m Machine, lastPtr LastPointer) (lcScanResult, LastPointer, bool) {
	var newestId int
	var newestCreateTime int64
	var latestPage []LifeCycleLog
	var logs []LifeCycleLog

	// target is the entry read in the previous cycle; nil on the first collection.
	var target *lcScanTarget
	if lastPtr.LcLastReadId != 0 {
		target = &lcScanTarget{
			Id:         lastPtr.LcLastReadId,
			createTime: lastPtr.LcLastReadCreateTime,
		}
	}

	bmcUrl := "https://" + m.BmcIP + c.rfLcPath
	seen := make(map[string]struct{})

	for page := 0; page < c.lcMaxPages; page++ {
		response, statusCode, err := c.fetchLifecycleLogPage(ctx, m, bmcUrl)

		lastPtr.LcLastHttpStatusCode = statusCode
		lastPtr.LcLastError = ""

		if err != nil {
			lastPtr.LcLastError = err.Error()
			return lcScanResult{}, lastPtr, false
		}

		if len(response.Members) == 0 {
			slog.Info("reached the end of the lifecycle log", "serial", m.Serial)
			return lcScanResult{
				logs:             logs,
				newestId:         newestId,
				newestCreateTime: newestCreateTime,
			}, lastPtr, true
		}

		if page == 0 {
			latestPage = response.Members
			id, ok := parseLifecycleLogId(m, response.Members[0])
			if !ok {
				return lcScanResult{}, lastPtr, false
			}
			newestId = id

			createTime, ok := parseLifecycleLogCreateTime(m, response.Members[0])
			if !ok {
				return lcScanResult{}, lastPtr, false
			}
			newestCreateTime = createTime.Unix()

			// Initial collection: collect only the latest page.
			if target == nil {
				return lcScanResult{
					logs:             response.Members,
					newestId:         newestId,
					newestCreateTime: newestCreateTime,
				}, lastPtr, true
			}

			// Log was cleared: collect only the latest page.
			if newestId < target.Id {
				slog.Warn("the lifecycle log was cleared in iDRAC; collecting from scratch", "serial", m.Serial, "lastReadId", target.Id, "newestId", newestId)
				return lcScanResult{
					logs:             response.Members,
					newestId:         newestId,
					newestCreateTime: newestCreateTime,
				}, lastPtr, true
			}
		}

		for _, v := range response.Members {
			id, ok := parseLifecycleLogId(m, v)
			if !ok {
				return lcScanResult{}, lastPtr, false
			}

			if id == target.Id && target.createTime != 0 {
				createTime, ok := parseLifecycleLogCreateTime(m, v)
				if !ok {
					return lcScanResult{}, lastPtr, false
				}

				// The same Id with a different creation time indicates that the log was cleared.
				if createTime.Unix() != target.createTime {
					slog.Warn("the lifecycle log was cleared in iDRAC; collecting from scratch", "serial", m.Serial, "lastReadId", target.Id, "Id", v.Id)
					return lcScanResult{
						logs:             latestPage,
						newestId:         newestId,
						newestCreateTime: newestCreateTime,
					}, lastPtr, true
				}
			}

			if id <= target.Id {
				return lcScanResult{
					logs:             logs,
					newestId:         newestId,
					newestCreateTime: newestCreateTime,
				}, lastPtr, true
			}

			// New entries added during pagination can shift pages and cause duplicates.
			if _, ok := seen[v.Id]; ok {
				continue
			}

			seen[v.Id] = struct{}{}
			logs = append(logs, v)
		}

		if response.NextLink == "" {
			slog.Warn("reached the end of the lifecycle log", "serial", m.Serial, "lastReadId", target.Id, "newestId", newestId)
			return lcScanResult{
				logs:             logs,
				newestId:         newestId,
				newestCreateTime: newestCreateTime,
			}, lastPtr, true
		}

		bmcUrl = "https://" + m.BmcIP + response.NextLink
	}

	counterLcCatchupTruncated.WithLabelValues(m.Serial).Inc()
	slog.Warn("stopped reading the lifecycle log at the page limit", "serial", m.Serial, "pageLimit", c.lcMaxPages, "lastReadId", target.Id, "newestId", newestId)
	return lcScanResult{
		logs:             logs,
		newestId:         newestId,
		newestCreateTime: newestCreateTime,
	}, lastPtr, true
}

func (c *logCollector) fetchLifecycleLogPage(ctx context.Context, m Machine, url string) (RedfishLcLogSchema, int, error) {
	byteJSON, statusCode, err := c.requestBmcLog(ctx, m, url, metricLogTypeLc)
	if err != nil {
		return RedfishLcLogSchema{}, statusCode, err
	}

	if statusCode != http.StatusOK {
		return RedfishLcLogSchema{}, statusCode, fmt.Errorf("unexpected HTTP status code: %d", statusCode)
	}

	var response RedfishLcLogSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		return RedfishLcLogSchema{}, statusCode, err
	}

	return response, statusCode, nil
}

// parseLifecycleLogId parses the numeric Id of an LC log entry.
func parseLifecycleLogId(m Machine, v LifeCycleLog) (int, bool) {
	id, err := strconv.Atoi(v.Id)
	if err != nil {
		slog.Error("failed to strconv; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Id", v.Id)
		return 0, false
	}

	return id, true
}

// parseLifecycleLogCreateTime parses the Created time of an LC log entry.
func parseLifecycleLogCreateTime(m Machine, v LifeCycleLog) (time.Time, bool) {
	createTime, err := time.Parse(time.RFC3339, v.Create)
	if err != nil {
		slog.Error("failed to parse for time; abort this cycle to keep the pointer unchanged", "err", err, "serial", m.Serial, "Created", v.Create)
		return time.Time{}, false
	}

	return createTime, true
}

// emitLifecycleLogs writes the entries, given newest first, in ascending order.
func (c *logCollector) emitLifecycleLogs(logs []LifeCycleLog, m Machine, logWriter bmcLogWriter) error {
	for _, v := range slices.Backward(logs) {
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
			slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog))
			return err
		}
	}

	return nil
}
