package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func doMainLoop(testMode bool, testModeConfig logCollector) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var testModeLoop int

	// setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// check parameter
	userId := os.Getenv("BMC_USER_ID")
	if len(userId) == 0 {
		slog.Error("The environment variable BMC_USER_ID should be set")
		return
	}

	password := os.Getenv("BMC_PASSWORD")
	if len(password) == 0 {
		slog.Error("The environment variable BMC_PASSWORD should be set")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cl := &http.Client{
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
	lc := logCollector{
		machinesPath: "/config/serverlist.json",
		rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:       "/data/pointers",
		rfClient:     cl,
		mutex:        &mu,
		user:         userId,
		password:     password,
		testMode:     false,
	}

	// test mode
	if testMode {
		lc.testMode = true
		lc.testOut = "testdata/output"
		lc.machinesPath = testModeConfig.machinesPath
		lc.ptrDir = testModeConfig.ptrDir
		lc.user = testModeConfig.user
		lc.password = testModeConfig.password
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	for {
		if testMode {
			testModeLoop++
			if testModeLoop > 3 {
				return
			}
		}
		select {
		case <-sigs: // Stop by Signal
			s := <-sigs
			err := fmt.Errorf("got signal %s", s)
			slog.Error(fmt.Sprintf("%s", err))
			return
		default:
			// get target machines list
			machinesList, err := machineListReader(lc.machinesPath)
			if err != nil {
				slog.Error(fmt.Sprintf("%s", err))
				cancel()
				return
			}
			// start log collector workers by BMC
			for i := 0; i < len(machinesList.Machine); i++ {
				wg.Add(1)
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
				// timer for stable running of go-routine
				time.Sleep(1 * time.Second)
			}
			wg.Wait()
		}
		// interval timer for next loop
		time.Sleep(10 * time.Second)
		lc.rfClient.CloseIdleConnections()
	}
}

func main() {
	doMainLoop(false, logCollector{})
}
