package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/tsr-transporter/bmc"
	"github.com/cybozu/neco-containers/tsr-transporter/dell"
	"github.com/cybozu/neco-containers/tsr-transporter/kintone"
	"github.com/cybozu/neco-containers/tsr-transporter/sabakan"
)

func doMain(bmc *bmc.UserConfig, sa *sabakan.Config, ka *kintone.Config) error {
	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)
	slog.Info("Start TSR transporter")

	// Signal treatment
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// Kintone
	App, err := kintone.NewKintoneEp(
		ka.Domain,
		ka.AppId,
		ka.SpaceId,
		ka.Guest,
		ka.Proxy,
		ka.Token)
	if err != nil {
		slog.Error("Error setting up the endpoint of Kintone app", "err", err)
		return err
	}
	httpStatus, records, err := App.CheckTsrRequest(ctx)
	if err != nil {
		slog.Error("Error accessing Kintone", "statusCode", httpStatus, "error", err)
		return err
	}
	if len(records.Record) == 0 {
		slog.Info("No exist TSR request")
	}

	// ここは Goルーチン化して、並列処理したい
	for _, record := range records.Record {
		// Retrive IPv4 by the serial
		serverSerial := record.Hostname.Value
		BmcIPv4, err := sabakan.GetBmcIpv4(sa.Ep, serverSerial)
		if err != nil {
			slog.Error("Error accessing Sabakan", "sabakan error", err)
			return err
		}
		// Obtain TSR from iDRAC
		d, err := dell.NewBmcEp(BmcIPv4, "support", bmc.Support.Password.Raw)
		if err != nil {
			slog.Error("Error accessing Sabakan")
			return err
		}
		job, err := d.StartCollection(ctx)
		if err != nil {
			slog.Error("Error start TSR job", "error", err)
			return err
		}
		err = d.WaitCollection(ctx, job)
		if err != nil {
			slog.Error("Error occurred while waiting JOB completion", "error", err)
			return err
		}
		FilenameTSR := fmt.Sprintf("%s-TSR.zip", serverSerial)
		f, err := os.Create(filepath.Join("/tmp", FilenameTSR))
		if err != nil {
			slog.Error("Error occurred when download file creation", "error", err)
			return err
		}
		defer f.Close()
		err = d.DownloadSupportAssist(ctx, f)
		if err != nil {
			slog.Error("Error occurred while downloading TSR", "error", err)
			return err
		}

		// Upload TSR to Kintone App
		var rec kintone.RecordWithFile
		rec.AppId = strconv.Itoa(App.AppId)
		rec.RecNum, _ = strconv.Atoi(record.RecordNumber.Value)
		rec.Recode.File.Value = make([]kintone.AttachedFile, 1)
		rec.Recode.File.Value[0].FileKey = ""
		rec.Recode.File.Value[0].Name = "/tmp/" + FilenameTSR /////////// これ修正が必要
		rec.Recode.TsrDate.Value = time.Now().Format(time.RFC3339)
		httpStatus, err = App.UploadFile(ctx, rec)
		if err != nil {
			slog.Error("Error occurred while uploading TSR", "error", err)
			return err
		}
		if httpStatus != 200 {
			slog.Error("Error occurred at HTTP level", "http status", httpStatus)
			return err
		}
	}
	slog.Info("Complete TSR transporter")
	return nil
}
