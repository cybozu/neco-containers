package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/neco-containers/tsr-transporter/bmc"
	"github.com/neco-containers/tsr-transporter/dell"
	"github.com/neco-containers/tsr-transporter/kintone"
	"github.com/neco-containers/tsr-transporter/sabakan"
	"github.com/spf13/pflag"
)

var (
	flgUserFile        *string = pflag.String("bmc-user-json", "/users/neco/bmc-user.json", "User and password of BMC")
	flgSabakanEndpoint *string = pflag.String("sabakan-ep", "https://api.sabakan.co.jp", "Sabakan endpoint")
	//flgKintoneEndpoint *string = pflag.String("kineton-ep", "https://api.kintone.co.jp", "Kintone endpoint")
	//flgKintoneAplToken *string = pflag.String("kintone-token", "secret-token", "token of kintone application")
	flgIntervalTime *int = pflag.Int("interval-time", 5, "Timer(sec) of checking interval time")
)

// 外部から渡せるようにしたい
type config struct {
	bmcUsername     string
	bmcPassword     string
	sabakanEndpoint string // URL
	kintone         *kintone.App
	intervalTime    time.Duration // interval (sec) time of scraping
}

func doMainLoop(cnf *config) {
	// Set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// Set interval time
	ticker := time.NewTicker(cnf.intervalTime)
	defer ticker.Stop()

	// Check loop
	var wg sync.WaitGroup
	var records kintone.Records
	var httpStatus int
	var err error
	for {
		select {
		case <-ctx.Done():
			slog.Error("stopped by", "signal", context.Cause(ctx))
			// Graceful stop when catch SIGTERM
			ticker.Stop()
			wg.Wait()
			return
		case <-ticker.C:
			// Kintone の新着チェック
			httpStatus, records, err = cnf.kintone.CheckReq(ctx)
			if err != nil {
				slog.Error("Error accessing Kintone", "statusCode", httpStatus, "error", err)
				return
			}

			// 複数のリクエストを取得するので、各レコードのループ処理が必要
			for _, record := range records.Record {
				fmt.Println("serial", record.Hostname.Value)

				// Sabakan シリアルチェック  URL, Serial
				serverSerial := record.Hostname.Value
				BmcIPv4, err := sabakan.GetBmcIpv4(cnf.sabakanEndpoint, serverSerial)
				if err != nil {
					slog.Error("Error accessing Sabakan", "sabakan error", err)
					return
				}
				// BMCからTSRを取得
				bmc, err := dell.NewBmcEp(BmcIPv4, cnf.bmcUsername, cnf.bmcPassword)
				if err != nil {
					slog.Error("Error accessing Sabakan")
					return
				}
				ctx := context.Background()
				job, err := bmc.StartCollection(ctx)
				if err != nil {
					slog.Error("Error start TSR job", "error", err)
					return
				}
				err = bmc.WaitCollection(ctx, job)
				if err != nil {
					slog.Error("Error occurred while waiting JOB completion", "error", err)
					return
				}
				FilenameTSR := fmt.Sprintf("%s-TSR.zip", serverSerial)
				f, err := os.Create(filepath.Join("/tmp", FilenameTSR))
				if err != nil {
					slog.Error("Error occurred when download file creation", "error", err)
					return
				}
				defer f.Close()

				err = bmc.DownloadSupportAssist(ctx, f)
				if err != nil {
					slog.Error("Error occurred while downloading TSR", "error", err)
					return
				}

				var recWithFile kintone.RecordWithFile
				recWithFile.AppId = strconv.Itoa(cnf.kintone.AppId)
				recWithFile.RecNum, _ = strconv.Atoi(record.RecordNumber.Value)
				recWithFile.Recode.File.Value = make([]kintone.AttachedFile, 1)
				recWithFile.Recode.File.Value[0].FileKey = ""
				recWithFile.Recode.File.Value[0].Name = "/tmp/" + FilenameTSR
				httpStatus, err = cnf.kintone.UploadFile(ctx, recWithFile)
				if err != nil {
					slog.Error("Error occurred while uploading TSR", "error", err)
					return
				}
				if httpStatus != 200 {
					slog.Error("Error occurred at HTTP level", "http status", httpStatus)
					// Kintoneからのエラーメッセージが欲しい
					return
				}
			}
		}
	}
}

func main() {
	pflag.Parse()

	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Read user & password for BMC
	userInfo, err := bmc.LoadBMCUserConfig(*flgUserFile)
	if err != nil {
		slog.Error("Can't read the user-list on BMC", "err", err)
		os.Exit(1)
	}

	//var App *kintone.App
	App, err := kintone.NewKintoneEp() // コンフィグファイルから読込む様に変更

	if err != nil {
		slog.Error("Error setting up the endpoint of Kintone app", "err", err)
		os.Exit(1)
	}

	config := config{
		bmcUsername:     "support",
		bmcPassword:     userInfo.Support.Password.Raw,
		sabakanEndpoint: *flgSabakanEndpoint,
		kintone:         App,
		intervalTime:    time.Duration(*flgIntervalTime) * time.Second,
	}

	slog.Info("TSR-requester started", "interval time", *flgUserFile)

	// Main Loop
	doMainLoop(&config)
}
