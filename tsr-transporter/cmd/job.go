package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cybozu/neco-containers/tsr-transporter/bmc"
	"github.com/cybozu/neco-containers/tsr-transporter/dell"
	"github.com/cybozu/neco-containers/tsr-transporter/kintone"
	"github.com/cybozu/neco-containers/tsr-transporter/sabakan"
	"github.com/spf13/cobra"
)

func replyErrKintoneRecord(id int, err error, a *kintone.App, r kintone.Fields, c context.Context) {
	var rec kintone.RecodeForUpdate
	rec.AppId = strconv.Itoa(a.AppId)
	rec.RecNum = id
	rec.Recode.Memo.Value = err.Error()
	rec.Recode.TsrDate.Value = time.Now().Format(time.RFC3339)
	sc, _, err := a.UpdateRecord(c, rec, http.MethodPut)
	if err != nil {
		slog.Error("Error failed writing a message to kintone", "error", err)
	}
	slog.Error("Error writing a message to kintone", "http status code", sc)
}

func acquireTsrPutKintone(bmc *bmc.UserConfig, sa *sabakan.Config, App *kintone.App, ctx context.Context, record kintone.Fields) error {
	// Retrive IPv4 by the serial
	serverSerial := record.Hostname.Value
	BmcIPv4, err := sabakan.GetBmcIpv4(sa.Ep, serverSerial)
	if err != nil {
		slog.Error("Error accessing Sabakan", "sabakan error", err)
		return err
	}
	if len(BmcIPv4) == 0 {
		// NOP when Not Found
		return nil
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
	fnTSR := fmt.Sprintf("%s-TSR.zip", serverSerial)
	fnFull := filepath.Join(App.WkDir, fnTSR)
	f, err := os.Create(fnFull)
	defer os.Remove(fnFull)
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
	rec.RecNum, err = strconv.Atoi(record.RecordNumber.Value)
	if err != nil {
		slog.Error("Error strconv.Atoi()", "error", err, "src string", record.RecordNumber.Value)
		return err
	}
	rec.Recode.File.Value = make([]kintone.AttachedFile, 1)
	rec.Recode.File.Value[0].FileKey = ""
	rec.Recode.File.Value[0].Name = fnFull
	rec.Recode.TsrDate.Value = time.Now().Format(time.RFC3339)
	httpStatus, err := App.UploadFile(ctx, rec)
	if err != nil {
		slog.Error("Error occurred while uploading TSR", "error", err)
		return err
	}
	if httpStatus != 200 {
		slog.Error("Error occurred at HTTP level", "http status", httpStatus)
		return err
	}
	return nil
}

func readConfiguration(cmd *cobra.Command, args []string) (*bmc.UserConfig, *sabakan.Config, *kintone.Config) {
	// iDRAC
	bc, err := bmc.ReadUsers(cfgBmcUser)
	if err != nil {
		slog.Error("Can't read the config file of BMC", "err", err)
		os.Exit(1)
	}

	// Sabakan
	sa, _ := sabakan.ReadConfig(cfgSabakan)
	if err != nil {
		slog.Error("Can't read the config file of sabakan", "err", err)
		os.Exit(1)
	}

	// Kintone
	ka, err := kintone.ReadAppConfig(cfgKintone)
	if err != nil {
		slog.Error("Can't read the config file of kintone", "err", err)
		os.Exit(1)
	}
	return bc, sa, ka
}

func jobMain(ctx context.Context, bmc *bmc.UserConfig, sa *sabakan.Config, ka *kintone.Config) error {
	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Setup endpoint of Kintone application
	App, err := kintone.NewKintoneEp(
		ka.Domain,
		ka.AppId,
		ka.SpaceId,
		ka.Guest,
		ka.Proxy,
		ka.Token,
		ka.WkDir)
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

	var wg sync.WaitGroup
	for _, record := range records.Record {
		wg.Add(1)
		go func() {
			slog.Info("Start TSR transporter")
			err := acquireTsrPutKintone(bmc, sa, App, ctx, record)
			if err != nil {
				id, _ := strconv.Atoi(record.RecordNumber.Value)
				// Write error message in Kintone Apl, ignore err in this function
				replyErrKintoneRecord(id, err, App, record, ctx)
			}
			wg.Done()
			slog.Info("Complete TSR transporter")
		}()
	}
	wg.Wait()
	return nil
}
