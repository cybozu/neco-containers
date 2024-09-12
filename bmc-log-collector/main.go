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

	// set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// set interval timer
	ticker := time.NewTicker(config.intervalTime)
	defer ticker.Stop()

	// expose metrics via HTTP
	go metrics("/metrics", ":8080")

	// scraping loop
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			slog.Error("stopped by", "signal", context.Cause(ctx))
			return
		case <-ticker.C:
			machinesList, err := readMachineList(config.machinesListDir)
			if err != nil {
				slog.Error("can't read the machine list", "err", err, "path", config.machinesListDir)
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

			// drop metrics which retired machine
			dropMetricsWhichRetiredMachine(machinesList)

			// remove ptr files that disappeared the serial in machineList
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
	// use default logger to prevent to mix log messages cross go-routine
	log.Print(stringJson)
	return nil
}

func dropMetricsWhichRetiredMachine(machinesList []Machine) {
	type retiredMachine struct {
		serial string
		nodeIP string
	}
	var dropList []retiredMachine

	for _, machine := range machinesList {
		if machine.State == "RETIRED" {
			dropList = append(dropList, retiredMachine{
				serial: machine.Serial,
				nodeIP: machine.NodeIP,
			})
		}
	}
	// delete by dropList entry
	for _, v := range dropList {
		deleteMetrics(v.serial, v.nodeIP)
	}
}

func main() {
	// setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// check parameter
	intervalTimeString := os.Getenv("BMC_INTERVAL_TIME")
	if len(intervalTimeString) == 0 {
		slog.Error("The environment variable BMC_INTERVAL_TIME should be set")
		os.Exit(1)
	}
	intervalTime, err := time.ParseDuration(intervalTimeString + "s")
	if err != nil {
		slog.Error("Can not convert string to time.Duration. please check second value")
		os.Exit(1)
	}

	// setup log scraping loop
	configLc := selCollector{
		userFile:        "/etc/neco/bmc-user.json",
		machinesListDir: "/config/serverlist.json",
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:          "/data/pointers",
		username:        "support",
		intervalTime:    intervalTime,
	}
	user, err := LoadBMCUserConfig(configLc.userFile)
	configLc.password = user.Support.Password.Raw
	if err != nil {
		slog.Error("Can't read the user-list on BMC")
		os.Exit(1)
	}

	// set BMC log writer
	logWriter := logProd{}
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	slog.Info("bmc-log-collector started", "interval time", intervalTime)
	doLogScrapingLoop(configLc, logWriter)
}
