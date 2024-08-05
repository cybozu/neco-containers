package main

import (
	"context"
	"crypto/tls"
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

func doLogScrapingLoop(testModeConfig selCollector, logWriter bmcLogWriter) {
	var wg sync.WaitGroup
	//var mu sync.Mutex
	var testModeLoop int

	flag := interface{}(logWriter)
	_, testMode := flag.(logTest)

	// check parameter
	username := os.Getenv("BMC_USERNAME")
	if len(username) == 0 {
		slog.Error("The environment variable BMC_USERNAME should be set")
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
	lc := selCollector{
		machinesListDir: "/config/serverlist.json",
		rfUriSel:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:          "/data/pointers",
		httpClient:      cl,
		//mutex:           &mu,
		username: username,
		password: password,
	}

	// test mode setup
	if testMode {
		lc.testOutputDir = "testdata/output"
		lc.machinesListDir = testModeConfig.machinesListDir
		lc.ptrDir = testModeConfig.ptrDir
		lc.username = testModeConfig.username
		lc.password = testModeConfig.password
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// set first interval timer
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// scraping loop
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
			return
		case <-ticker.C:
			machinesList, err := machineListReader(lc.machinesListDir)
			if err != nil {
				slog.Error("machineListReader()", "err", err, "path", lc.machinesListDir)
				return
			}
			// start log collector workers by BMCs
			for _, m := range machinesList {
				wg.Add(1)
				go func() {
					//lc.logCollectorWorker(ctx, &wg, m, logWriter)
					lc.selCollectorWorker(ctx, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
			lc.httpClient.CloseIdleConnections()
		}
		// scrape cycle: 30min (=1800sec)
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

	// setup BMC log writer
	logWriter := logProd{}

	// log scraping loop
	doLogScrapingLoop(selCollector{}, logWriter)
}
