package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"sync"
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

type Redfish struct {
	Name        string           `json:"Name"`
	Count       int              `json:"Members@odata.count"`
	Context     string           `json:"@odata.context"`
	Id          string           `json:"@odata.id"`
	Type        string           `json:"@odata.type"`
	Description string           `json:"Descriptionta"`
	Sel         []SystemEventLog `json:"Members"`
}

type logCollector struct {
	machinesPath string          // Machine list path
	rfUrl        string          // Redfish API address
	ptrDir       string          // Pointer store
	wg           *sync.WaitGroup // wait ????????????????
	testMode     bool            // when testMode is true, write text file for test
	testOut      string          // Test output directory
}

func (c *logCollector) worker(ctx context.Context, m Machine) {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			byteBuf, err := bmcClient("https://" + m.BmcIP + c.rfUrl)
			if err != nil {
				errmsg := fmt.Sprintf("%s", err)
				slog.Error(errmsg)
			}
			c.antiDuplicatefilter(byteBuf, m, c.ptrDir)
			return
		}
	}
}

func (c *logCollector) antiDuplicatefilter(byteJSON []byte, server Machine, ptrDir string) {

	var members Redfish
	if err := json.Unmarshal(byteJSON, &members); err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return
	}

	lastPtr, err := readLastPointer(server.Serial, ptrDir)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
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
			v, _ := json.Marshal(members.Sel[i])
			fmt.Println(string(v))
			lastPtr.LastReadId = lastId
			lastPtr.LastReadTime = createUnixtime
			if c.testMode {
				testPrint(c.testOut, server.Serial, string(v))
			}
		} else if lastPtr.LastReadId > lastId {
			// ID set to 1 with iDRAC log clear by WebUI, should compare with both its unixtime to identify the latest.
			if lastPtr.LastReadTime < createUnixtime {
				v, _ := json.Marshal(members.Sel[i])
				fmt.Println(string(v))
				lastPtr.LastReadId = lastId
				lastPtr.LastReadTime = createUnixtime
				if c.testMode {
					testPrint(c.testOut, server.Serial, string(v))
				}
			}
		}
	}

	err = updateLastPointer(LastPointer{
		Serial:       server.Serial,
		LastReadTime: createUnixtime,
		LastReadId:   lastId,
	}, ptrDir)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return
	}
}

func testPrint(dir string, serial string, output string) {
	fn := path.Join(dir, serial)
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return
	}
	defer file.Close()
	file.WriteString(fmt.Sprintln(string(output)))
}
