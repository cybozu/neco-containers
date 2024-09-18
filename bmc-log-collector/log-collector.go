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
	const layout = "2006-01-02T15:04:05Z07:00"
	var createUnixtime int64
	var lastId int
	filePath := path.Join(c.ptrDir, m.Serial)

	err := checkAndCreatePointerFile(filePath)
	if err != nil {
		slog.Error("can't check a pointer file.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	lastPtr, err := readLastPointer(filePath)
	if err != nil {
		slog.Error("can't read a pointer file.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	bmcUrl := "https://" + m.BmcIP + c.rfSelPath
	byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, bmcUrl)
	if statusCode != 200 || err != nil {
		// Increment the failed counter
		counterRequestFailed.WithLabelValues(m.Serial).Inc()

		// Prevent log output by the same error code or httpStatus
		if statusCode != 200 && statusCode != lastPtr.LastHttpStatusCode {
			slog.Error("failed access to iDRAC on HTTP level.", "url", bmcUrl, "httpStatusCode", statusCode)
		}
		lastPtr.LastHttpStatusCode = statusCode

		if err != nil && lastPtr.LastError != err {
			slog.Error("failed access to iDRAC at TCP/IP level.", "err", err, "BMC IP", m.BmcIP)
		}
		lastPtr.LastError = err

		err = updateLastPointer(lastPtr, filePath)
		if err != nil {
			slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
		}
		return
	}

	// Increment the success counter
	counterRequestSuccess.WithLabelValues(m.Serial).Inc()

	var members RedfishJsonSchema
	if err := json.Unmarshal(byteJSON, &members); err != nil {
		slog.Error("failed to translate JSON to go struct.", "err", err, "serial", m.Serial, "ptrDir", c.ptrDir)
		return
	}

	for i := len(members.Sel) - 1; i >= 0; i-- {
		t, _ := time.Parse(layout, members.Sel[i].Create)
		createUnixtime = t.Unix()
		lastId, _ = strconv.Atoi(members.Sel[i].Id)
		members.Sel[i].Serial = m.Serial
		members.Sel[i].BmcIP = m.BmcIP
		members.Sel[i].NodeIP = m.NodeIP

		if lastPtr.LastReadId < lastId {
			// Normal case without iDRAC log reset.
			bmcByteJsonLog, err := json.Marshal(members.Sel[i])
			if err != nil {
				slog.Error("failed to marshal the system event log", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
			}
			err = logWriter.write(string(bmcByteJsonLog), m.Serial)
			if err != nil {
				slog.Error("failed to output log", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
			}
			lastPtr.LastReadId = lastId
			lastPtr.LastReadTime = createUnixtime
		} else {
			// If the log is reset in iDRAC, the ID starts from 1.
			// In that case, determine if the log has already been
			// issued based on the time it was generated.
			if lastPtr.LastReadTime < createUnixtime {
				bmcByteJsonLog, err := json.Marshal(members.Sel[i])
				if err != nil {
					slog.Error("failed to convert JSON", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
				}
				err = logWriter.write(string(bmcByteJsonLog), m.Serial)
				if err != nil {
					slog.Error("failed to output log", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
				}
				lastPtr.LastReadId = lastId
				lastPtr.LastReadTime = createUnixtime
			}
		}
	}

	err = updateLastPointer(LastPointer{
		LastReadTime: createUnixtime,
		LastReadId:   lastId,
		LastError:    nil,
	}, filePath)
	if err != nil {
		slog.Error("failed to write a pointer file.", "err", err, "serial", m.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", c.ptrDir)
		return
	}
}
