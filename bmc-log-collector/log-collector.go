package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"
)

type SystemEventLog struct {
	Od_Id             string   `json:"@odata.id"`
	Od_Type           string   `json:"@odata.type"`
	Create            string   `json:"Created"`
	Description       string   `json:"Description"`
	EntryCode         string   `json:"EntryCode"`
	EntryType         string   `json:"EntryType"`
	GeneratorId       string   `json:"GeneratorId"`
	Id                string   `json:"Id"`
	Message           string   `json:"Message"`
	MessageArgs       []string `json:"MessageArgs"`
	OdCnt_MessageArgs int      `json:"MessageArgs@odata.count"`
	MessageId         string   `json:"MessageId"`
	Name              string   `json:"Name"`
	SensorNumber      int      `json:"SensorNumber"`
	SensorType        string   `json:"SensorType"`
	Severity          string   `json:"Severity"`
	Serial            string
	NodeIP            string
	BmcIP             string
}

type RedfishJsonSchema struct {
	Name        string           `json:"Name"`
	Count       int              `json:"Members@odata.count"`
	Context     string           `json:"@odata.context"`
	Id          string           `json:"@odata.id"`
	Type        string           `json:"@odata.type"`
	Description string           `json:"Descriptionta"`
	Sel         []SystemEventLog `json:"Members"`
}

// SEL(System Event Log) Collector
type selCollector struct {
	machinesListDir string        // Directory of the machines list
	rfSelPath       string        // SEL path of Redfish API address
	ptrStore        *pointerStore // Pointers of each machine
	startupLookback time.Duration // On the first read of a machine, output only entries created within this duration
	username        string        // iDRAC username
	password        string        // iDRAC password
	httpClient      *http.Client  // to reuse HTTP transport
	intervalTime    time.Duration // interval (sec) time of scraping
}

func (c *selCollector) collectSystemEventLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	lastPtr := c.ptrStore.get(m.Serial)
	requestTime := time.Now()

	bmcUrl := "https://" + m.BmcIP + c.rfSelPath
	byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, bmcUrl)
	if err != nil {
		// Increment the failed counter
		counterRequestFailed.WithLabelValues(m.Serial).Inc()
		// Prevent log output by the same error code
		if lastPtr.LastError != err.Error() {
			slog.Error("failed access to iDRAC on TCP/IP level.", "url", bmcUrl, "err", err.Error(), "serial", m.Serial)
		}
		lastPtr.LastHttpStatusCode = 0
		lastPtr.LastError = err.Error()
		c.ptrStore.put(m.Serial, lastPtr)
		return
	}
	if statusCode != 200 {
		// Increment the failed counter
		counterRequestFailed.WithLabelValues(m.Serial).Inc()
		// Prevent log output by the same httpStatus
		if statusCode != lastPtr.LastHttpStatusCode {
			slog.Error("failed access to iDRAC on HTTP level.", "url", bmcUrl, "httpStatusCode", statusCode, "serial", m.Serial)
		}
		lastPtr.LastHttpStatusCode = statusCode
		lastPtr.LastError = ""
		c.ptrStore.put(m.Serial, lastPtr)
		return
	}

	// Increment the success counter
	counterRequestSuccess.WithLabelValues(m.Serial).Inc()

	var response RedfishJsonSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial)
		return
	}

	// some iDRAC version can't retrieve SEL
	if len(response.Sel) == 0 {
		slog.Error("can not retrieve system event log", "SEL entry", 0, "serial", m.Serial)
		return
	}

	createTime, err := time.Parse(time.RFC3339, response.Sel[len(response.Sel)-1].Create)
	if err != nil {
		slog.Error("failed to parse for time", "err", err, "serial", m.Serial)
		return
	}
	firstCreateTime := createTime.Unix()

	for i, v := range slices.Backward(response.Sel) {
		currentId, err := strconv.Atoi(v.Id)
		if err != nil {
			slog.Error("failed to strconv", "err", err, "serial", m.Serial, "LastReadId", currentId)
			continue
		}
		// Add the information to identify of the node
		v.Serial = m.Serial
		v.BmcIP = m.BmcIP
		v.NodeIP = m.NodeIP

		// The pointer is not persisted, so on the first read after start, output only recent entries.
		// Duplicates and losses around restart are acceptable.
		if !lastPtr.SelRead {
			if c.isRecent(v, requestTime) {
				bmcByteJsonLog, err := json.Marshal(v)
				if err != nil {
					slog.Error("failed to marshal the system event log", "err", err, "serial", m.Serial, "currentLastReadId", currentId)
				}
				err = logWriter.write(string(bmcByteJsonLog), m.Serial)
				if err != nil {
					slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", currentId)
				}
			}
			lastPtr.LastReadId = currentId
			continue
		}

		if lastPtr.LastReadId < currentId {
			bmcByteJsonLog, err := json.Marshal(v)
			if err != nil {
				slog.Error("failed to marshal the system event log", "err", err, "serial", m.Serial, "lastPtr.LastReadId", lastPtr.LastReadId, "currentLastReadId", currentId)
			}

			err = logWriter.write(string(bmcByteJsonLog), m.Serial)
			if err != nil {
				slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", currentId)
			}

			lastPtr.LastReadId = currentId
		} else {
			// If the log is reset in iDRAC, the ID starts from 1.
			// In that case, determine if generated time been changed to identify log reseted.
			if lastPtr.FirstCreateTime != firstCreateTime {
				bmcByteJsonLog, err := json.Marshal(v)
				if err != nil {
					slog.Error("failed to convert JSON", "err", err, "serial", m.Serial, "i", i, "Event", v, "currentLastReadId", currentId)
				}

				// Output duplicate log, after log reset in iDRAC
				err = logWriter.write(string(bmcByteJsonLog), m.Serial)
				if err != nil {
					slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", currentId)
				}

				lastPtr.LastReadId = currentId
			}
		}
	}

	lastPtr.FirstCreateTime = firstCreateTime
	lastPtr.SelRead = true
	c.ptrStore.put(m.Serial, lastPtr)
}

// isRecent reports whether the entry was created within the startup lookback before requestTime.
// An entry whose created time can't be parsed is treated as recent.
func (c *selCollector) isRecent(v SystemEventLog, requestTime time.Time) bool {
	created, err := time.Parse(time.RFC3339, v.Create)
	if err != nil {
		return true
	}
	return !created.Before(requestTime.Add(-c.startupLookback))
}
