package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path"
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
	ptrDir          string        // Pointer store
	username        string        // iDRAC username
	password        string        // iDRAC password
	httpClient      *http.Client  // to reuse HTTP transport
	intervalTime    time.Duration // interval (sec) time of scraping
}

func (c *selCollector) collectSystemEventLog(ctx context.Context, m Machine, logWriter bmcLogWriter) {
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
		err = updateLastPointer(lastPtr, filePath)
		if err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		}
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
		err = updateLastPointer(lastPtr, filePath)
		if err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "filePath", filePath)
		}
		return
	}

	// Increment the success counter
	counterRequestSuccess.WithLabelValues(m.Serial).Inc()

	var response RedfishJsonSchema
	if err := json.Unmarshal(byteJSON, &response); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	createTime, err := time.Parse(time.RFC3339, response.Sel[len(response.Sel)-1].Create)
	if err != nil {
		slog.Error("failed to parse for time", "err", err, "serial", m.Serial)
		return
	}
	firstCreateTime := createTime.Unix()

	for i := len(response.Sel) - 1; i >= 0; i-- {
		currentId, err := strconv.Atoi(response.Sel[i].Id)
		if err != nil {
			slog.Error("failed to strconv", "err", err, "serial", m.Serial, "LastReadId", currentId, "ptrDir", c.ptrDir)
			continue
		}
		// Add the information to identify of the node
		response.Sel[i].Serial = m.Serial
		response.Sel[i].BmcIP = m.BmcIP
		response.Sel[i].NodeIP = m.NodeIP

		if lastPtr.LastReadId < currentId {
			bmcByteJsonLog, err := json.Marshal(response.Sel[i])
			if err != nil {
				slog.Error("failed to marshal the system event log", "err", err, "serial", m.Serial, "lastPtr.LastReadId", lastPtr.LastReadId, "currentLastReadId", currentId, "ptrDir", c.ptrDir)
			}

			err = logWriter.write(string(bmcByteJsonLog), m.Serial)
			if err != nil {
				slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", currentId, "ptrDir", c.ptrDir)
			}

			lastPtr.LastReadId = currentId
			lastPtr.LastError = ""
		} else {
			// If the log is reset in iDRAC, the ID starts from 1.
			// In that case, determine if generated time been changed to identify log reseted.
			if lastPtr.FirstCreateTime != firstCreateTime {
				bmcByteJsonLog, err := json.Marshal(response.Sel[i])
				if err != nil {
					slog.Error("failed to convert JSON", "err", err, "serial", m.Serial, "i", i, "Event", response.Sel[i], "currentLastReadId", currentId)
				}

				err = logWriter.write(string(bmcByteJsonLog), m.Serial)
				if err != nil {
					slog.Error("failed to output log", "err", err, "serial", m.Serial, "bmcByteJsonLog", string(bmcByteJsonLog), "currentLastReadId", currentId)
				}
				lastPtr.LastReadId = currentId
				lastPtr.LastError = ""
			}
		}
	}
	lastPtr.FirstCreateTime = firstCreateTime
	err = updateLastPointer(lastPtr, filePath)
	if err != nil {
		slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "firstCreateTime", firstCreateTime, "filePath", filePath)
		return
	}
}
