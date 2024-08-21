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
)

type bmcLogWriter interface {
	write(stringJson string, serial string) (err error)
}

func doLogScrapingLoop(config selCollector, logWriter bmcLogWriter) {
	var wg sync.WaitGroup
	config.httpClient = &http.Client{
		Timeout: time.Duration(10) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 20 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// set interval timer
	ticker := time.NewTicker(config.intervalTime * time.Second)
	defer ticker.Stop()

	// Expose metrics via HTTP
	go metrics("/metrics", ":8080")

	// scraping loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop by signal~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			//s := <-sigs
			//slog.Info("ctx.Done", "Signal", s.String()) 代替手段を調べる ctxで出せるかも
			return
		case <-ticker.C:
			machinesList, err := readMachineList(config.machinesListDir)
			if err != nil {
				slog.Error("machineListReader()", "err", err, "path", config.machinesListDir)
				return
			}
			// start log collector workers by BMCs
			for _, m := range machinesList {
				wg.Add(1)
				go func() {
					config.collectSystemEventLog(ctx, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
		}

		// Remove ptr files that no update for 6 months
		err := deleteUnUpdatedFiles(config.ptrDir)
		if err != nil {
			slog.Error("deleteUnUpdatedFiles()", "err", err, "path", config.ptrDir)
		}
	}
}

// BMC log writer to forward Loki
type logProd struct{}

func (l logProd) write(stringJson string, serial string) error {
	// use default logger to prevent to mix log messages cross go-routine
	log.Print(stringJson)
	return nil
}

func main() {

	// check parameter
	username := os.Getenv("BMC_USERNAME")
	if len(username) == 0 {
		slog.Error("The environment variable BMC_USERNAME should be set")
		os.Exit(1)
	}

	password := os.Getenv("BMC_PASSWORD")
	if len(password) == 0 {
		slog.Error("The environment variable BMC_PASSWORD should be set")
		os.Exit(1)
	}

	// setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// setup log scraping loop
	configLc := selCollector{
		machinesListDir: "/config/serverlist.json",
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:          "/data/pointers",
		username:        username,
		password:        password,
		intervalTime:    1800,
	}

	// set BMC log writer
	logWriter := logProd{}
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	doLogScrapingLoop(configLc, logWriter)
}
