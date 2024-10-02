package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

type bmcLogWriter interface {
	write(stringJson string, serial string) (err error)
}

func doLogScrapingLoop(config selCollector, logWriter bmcLogWriter) {
	config.httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 20 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
		},
	}

	// Set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// Set interval time
	ticker := time.NewTicker(config.intervalTime)
	defer ticker.Stop()

	// Expose metrics via HTTP
	go metrics("/metrics", ":8080")

	// Scraping loop
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			slog.Error("stopped by", "signal", context.Cause(ctx))
			// Graceful stop when catch SIGTERM
			ticker.Stop()
			wg.Wait()
			return
		case <-ticker.C:
			machinesList, err := readMachineList(config.machinesListDir)
			if err != nil {
				slog.Error("can't read the machine list", "err", err, "path", config.machinesListDir)
				return
			}
			// Start log collector workers by BMCs
			for _, m := range machinesList {
				wg.Add(1)
				go func() {
					config.collectSystemEventLog(ctx, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
			// Remove ptr files that disappeared the serial in machineList
			err = deletePtrFileDisappearedSerial(config.ptrDir, machinesList)
			if err != nil {
				slog.Error("failed remove the pointer file", "err", err, "path", config.ptrDir)
			}
		}
	}
}

// BMC log writer to forward Loki
type logProd struct{}

func (l logProd) write(stringJson string, serial string) error {
	// Use default logger to prevent to mix log messages cross go-routine
	log.Print(stringJson)
	return nil
}

var (
	flgUserFile             *string = pflag.String("bmc-user-json", "/users/neco/bmc-user.json", "User and password of BMC")
	flgUserId               *string = pflag.String("user-id", "support", "User ID of bmc-user-json JSON file")
	flgMachineList          *string = pflag.String("machine-list-json", "/config/machineslist.json", "Target machines list of log scraping")
	flgPointerDir           *string = pflag.String("pointer-dir-path", "/data/pointers", "Data directory of pointer management")
	flgScrapingIntervalTime *int    = pflag.Int("scraping-interval-time", 300, "Timer(sec) of scraping interval time")
)

func main() {
	pflag.Parse()

	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Read user & password for BMC
	user, err := LoadBMCUserConfig(*flgUserFile)
	if err != nil {
		slog.Error("Can't read the user-list on BMC", "err", err)
		os.Exit(1)
	}

	// Setup log scraping loop
	configLc := selCollector{
		machinesListDir: *flgMachineList,
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:          *flgPointerDir,
		username:        *flgUserId,
		password:        user.Support.Password.Raw,
		intervalTime:    time.Duration(*flgScrapingIntervalTime) * time.Second,
	}

	// Set BMC log writer
	logWriter := logProd{}
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	slog.Info("bmc-log-collector started", "interval time", *flgScrapingIntervalTime)
	doLogScrapingLoop(configLc, logWriter)
}
