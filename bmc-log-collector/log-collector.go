package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
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
	machinesListDir string       // Directory of the machines list
	rfUriSel        string       // SEL path of Redfish API address
	ptrDir          string       // Pointer store
	testOutputDir   string       // Test output directory
	username        string       // iDRAC username
	password        string       // iDRAC password
	httpClient      *http.Client // to reuse HTTP transport
}

func (c *selCollector) selCollectorWorker(ctx context.Context, m Machine, logWriter bmcLogWriter) {
	bmcUrl := "https://" + m.BmcIP + c.rfUriSel
	byteJSON, err := requestToBmc(ctx, c.username, c.password, c.httpClient, bmcUrl)

	if err != nil {
		// When canceled by context, the pointer files is never updated because the return is made here.
		slog.Error("requestToBmc()", "err", err, "url", c.rfUriSel)
		return
	}
	c.bmcLogOutputWithoutDuplication(byteJSON, m, c.ptrDir, logWriter)
}

func (c *selCollector) bmcLogOutputWithoutDuplication(byteJSON []byte, server Machine, ptrDir string, logWriter bmcLogWriter) {

	var members RedfishJsonSchema
	if err := json.Unmarshal(byteJSON, &members); err != nil {
		slog.Error("json.Unmarshal()", "err", err, "serial", server.Serial, "ptrDir", ptrDir)
		return
	}

	lastPtr, err := readLastPointer(server.Serial, ptrDir)
	if err != nil {
		slog.Error("readLastPointer()", "err", err, "serial", server.Serial, "ptrDir", ptrDir)
		return
	}

	layout := "2006-01-02T15:04:05Z07:00"
	var createUnixtime int64
	var lastId int

	for i := len(members.Sel) - 1; i >= 0; i-- {
		t, _ := time.Parse(layout, members.Sel[i].Create)
		createUnixtime = t.Unix()
		lastId, _ = strconv.Atoi(members.Sel[i].Id)
		members.Sel[i].Serial = server.Serial
		members.Sel[i].BmcIP = server.BmcIP
		members.Sel[i].NodeIP = server.NodeIP

		// Anti duplication
		if lastPtr.LastReadId < lastId {
			bmcByteJsonLog, _ := json.Marshal(members.Sel[i])
			logWriter.writer(string(bmcByteJsonLog), server.Serial)
			lastPtr.LastReadId = lastId
			lastPtr.LastReadTime = createUnixtime
		} else if lastPtr.LastReadId > lastId {
			// ID set to 1 with iDRAC log clear by WebUI, should compare with both its unixtime to identify the latest.
			if lastPtr.LastReadTime < createUnixtime {
				bmcByteJsonLog, _ := json.Marshal(members.Sel[i])
				logWriter.writer(string(bmcByteJsonLog), server.Serial)
				lastPtr.LastReadId = lastId
				lastPtr.LastReadTime = createUnixtime
			}
		}
	}

	err = updateLastPointer(LastPointer{
		Serial:         server.Serial,
		LastReadTime:   createUnixtime,
		LastReadId:     lastId,
		LastUpdateTime: time.Now().Unix(),
	}, ptrDir)

	if err != nil {
		slog.Error("updateLastPointer()", "err", err, "serial", server.Serial, "createUnixtime", createUnixtime, "LastReadId", lastId, "ptrDir", ptrDir)
		return
	}
}
