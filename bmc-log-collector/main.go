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

func doLogScrapingLoop(config logCollector, logWriter bmcLogWriter) {
	config.httpClient = &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 60 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 60 * time.Second,
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
			// Start log collector workers by BMCs.
			// Collect the logs sequentially in a worker to avoid
			// concurrent accesses to the same iDRAC.
			for _, m := range machinesList {
				wg.Go(func() {
					config.collectSystemEventLog(ctx, m, logWriter)
					config.collectLifecycleLog(ctx, m, logWriter)
				})
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
	flgLcMaxPages           *int    = pflag.Int("lclog-max-pages", 10, "Maximum pages of the lifecycle log to read per scraping cycle")
)

func main() {
	pflag.Parse()

	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	if *flgLcMaxPages < 1 {
		slog.Error("lclog-max-pages must be 1 or larger", "lclog-max-pages", *flgLcMaxPages)
		os.Exit(1)
	}
	if *flgScrapingIntervalTime < 1 {
		slog.Error("scraping-interval-time must be 1 or larger", "scraping-interval-time", *flgScrapingIntervalTime)
		os.Exit(1)
	}

	// Read user & password for BMC
	user, err := LoadBMCUserConfig(*flgUserFile)
	if err != nil {
		slog.Error("Can't read the user-list on BMC", "err", err)
		os.Exit(1)
	}

	// Setup log scraping loop
	configLc := logCollector{
		machinesListDir: *flgMachineList,
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		rfLcPath:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Lclog/Entries",
		ptrDir:          *flgPointerDir,
		username:        *flgUserId,
		password:        user.Support.Password.Raw,
		intervalTime:    time.Duration(*flgScrapingIntervalTime) * time.Second,
		lcMaxPages:      *flgLcMaxPages,
	}

	// Set BMC log writer
	logWriter := logProd{}
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	slog.Info("bmc-log-collector started", "interval time", *flgScrapingIntervalTime)
	doLogScrapingLoop(configLc, logWriter)
}
