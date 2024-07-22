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
	machinesPath string          // ログ収集対象マシンのPath
	miniNum      int             // 最小ワーカー数
	maxiNum      int             // 最大ワーカー数
	currNum      int             // 現在のワーカー数
	rfUrl        string          // Redfishのパス
	ptrDir       string          // ポインタ用ディレクトリ
	ctx          context.Context // コンテキスト
	cancel       context.CancelFunc
	que          MessageQueue    // マシンのキュー
	interval     time.Duration   // サイクル間の待機秒数
	wg           *sync.WaitGroup // 待ち合わせ用
	testMode     bool            // テストモード 出力をファイルへ出す
	testOut      string          // テスト用出力先
}

func (c *logCollector) worker(n int) {
	slog.Info(fmt.Sprintf("log-collector %d started", n))

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			slog.Info("log-collectors stopped")
			return
		default:
			targetMachine := c.que.get2()
			slog.Info(fmt.Sprintf("Worker %d", n))
			byteBuf, err := bmcClient("https://" + targetMachine.BmcIP + c.rfUrl)
			if err != nil {
				errmsg := fmt.Sprintf("%s", err)
				slog.Error(errmsg)
			}
			c.printLogs(byteBuf, targetMachine, c.ptrDir)
			// Interval timer
			time.Sleep(c.interval * time.Second)
		}
	}
}

func (c *logCollector) printLogs(byteJSON []byte, server Machine, ptrDir string) {

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

		// IDの大小で比較して出力 クリアでId=1に戻った時はシリアル時刻の大小で比較
		if lastPtr.LastReadId < lastId {
			v, _ := json.Marshal(members.Sel[i])
			fmt.Println(string(v))
			lastPtr.LastReadId = lastId
			lastPtr.LastReadTime = createUnixtime
			if c.testMode {
				testPrint(c.testOut, server.Serial, string(v))
			}
		} else if lastPtr.LastReadId > lastId {
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
