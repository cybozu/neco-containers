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

	// set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// set interval timer
	ticker := time.NewTicker(config.intervalTime * time.Second)
	defer ticker.Stop()

	// expose metrics via HTTP
	go metrics("/metrics", ":8080")

	// scraping loop
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
			err = dropMetricsWhichRetiredMachine(config.ptrDir, machinesList)
			if err != nil {
				slog.Error("failed to drop metrics", "err", err, "pointer directory", config.ptrDir)
			}
		}

		// remove ptr files that no update for 6 months
		err := deleteUnUpdatedFiles(config.ptrDir)
		if err != nil {
			slog.Error("failed remove pointer file which did not update for 6 month.", "err", err, "path", config.ptrDir)
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

func dropMetricsWhichRetiredMachine(ptrDir string, machinesList []Machine) error {
	type retiredMachine struct {
		serial string
		nodeIP string
	}

	// 連想配列に変換
	x := make(map[string]bool, len(machinesList))
	for _, m := range machinesList {
		x[m.Serial] = true
	}

	// 過去のアクセス記録からリストを取得
	// 連想配列を返せないか？
	machines, err := getMachineListWhichEverAccessed(ptrDir)
	if err != nil {
		return err
	}

	var dropList []retiredMachine
	// append で削除リストを作るのが良い
	for _, server := range machines {
		_, isExist := x[server.Serial]
		fmt.Printf("check isExist %v\n", isExist)
		//fmt.Println("============================= key", server.Serial, "val", machines[server.Serial])
		if !isExist {
			dropList = append(dropList, retiredMachine{
				serial: server.Serial,
				nodeIP: server.NodeIP,
			})
		}
	}
	fmt.Println("================================== Drop List", dropList)

	// 削除リストにあるものを削除
	for _, v := range dropList {
		deleteMetrics(v.serial, v.nodeIP)
	}
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
