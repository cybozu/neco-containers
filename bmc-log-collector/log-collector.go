package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	que          Queue           // マシンのキュー
	interval     time.Duration   // サイクル間の待機秒数
	wg           *sync.WaitGroup // 待ち合わせ用
}

func (c *logCollector) worker(n int) {

	fmt.Printf("collector #%d Started\n", n)

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			slog.Error("log-collectors stopped")
			return
		default:
			targetMachine := c.que.get()
			slog.Debug(fmt.Sprintf("Worker %d", n))

			// タイムアウトが必要
			byteBuf, err := bmcClient("https://" + targetMachine.BmcIP + c.rfUrl)
			if err != nil {
				slog.Error("Error")
			}
			// unmarshal log record
			// 下のレイヤーでプリントしないで、いったん、戻して、アウトプットしては？
			// 画面に出すか？ファイルへ出すか選択できると良い？
			printLogs(byteBuf, targetMachine, c.ptrDir)

		}
	}
}

func printLogs(byteJSON []byte, server Machine, ptrDir string) {

	var members Redfish
	if err := json.Unmarshal(byteJSON, &members); err != nil {
		slog.Error("failed to convert struct from JSON")
		return
	}

	lastPtr, err := readLastPointer(server.Serial, ptrDir)
	if err != nil {
		slog.Error("failed to get last log pointer")
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
		} else if lastPtr.LastReadId > lastId {
			if lastPtr.LastReadTime < createUnixtime {
				v, _ := json.Marshal(members.Sel[i])
				fmt.Println(string(v))
			}
		}
	}

	err = updateLastPointer(LastPointer{
		Serial:       server.Serial,
		LastReadTime: createUnixtime,
		LastReadId:   lastId,
	}, ptrDir)
	if err != nil {
		//slog.Error("failed to update log pointer")
		return
	}
}
