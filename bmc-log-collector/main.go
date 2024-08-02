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
	writer(stringJson string, serial string) (err error)
}

func doMainLoop(testModeConfig logCollector, logWriter bmcLogWriter) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var testModeLoop int

	flag := interface{}(logWriter)
	_, testMode := flag.(logTest)

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
		httpClient:   cl,
		mutex:        &mu,
		user:         userId,
		password:     password,
		//testMode:     false,
	}

	// test mode
	if testMode {
		//lc.testMode = true
		lc.testOut = "testdata/output"
		lc.machinesPath = testModeConfig.machinesPath
		lc.ptrDir = testModeConfig.ptrDir
		lc.user = testModeConfig.user
		lc.password = testModeConfig.password
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// set first interval timer
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// main loop
	for {
		// when use the test mode, must break infinite loop
		if testMode {
			testModeLoop++
			if testModeLoop > 3 {
				return
			}
		}
		select {
		case <-ctx.Done():
			//cancel()
			return
		case <-ticker.C:
			machinesList, err := machineListReader(lc.machinesPath)
			if err != nil {
				slog.Error("machineListReader()", "err", err, "path", lc.machinesPath)
				return
			}
			// start log collector workers by BMC
			//for i := 0; i < len(machinesList.Machine); i++ {
			for _, m := range machinesList.Machine {
				wg.Add(1)
				go func() {
					lc.logCollectorWorker(ctx, &wg, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
			lc.httpClient.CloseIdleConnections()
		}
		// scrape cycle: 30min (=1800sec)
		fmt.Println("*********** waiting ********************")
		intervalTime := 1800 * time.Second
		if testMode {
			intervalTime = 10 * time.Second
		}
		ticker = time.NewTicker(intervalTime)
	}
}

// BMC log writer to forward Loki via promtail
type logProd struct{}

func (l logProd) writer(stringJson string, serial string) error {
	// use default logger to prevent to mix log messages cross go-routine
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.Print(stringJson)
	return nil
}

func main() {
	// setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// set BMC log writer
	logWriter := logProd{}
	doMainLoop(logCollector{}, logWriter)
}
