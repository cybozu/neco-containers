package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Log collector main loop
func main() {


	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cl := &http.Client{
		Timeout:   time.Duration(10) * time.Second,
		Transport: tr,
	}
	lc := logCollector{
		machinesPath: "testdata/conf/serverlist.json",
		rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:       "pointers",
		rfclient:     cl,
		testMode:     false,
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	for {
		select {
		// Stop by Signal
		case <-sigs:
			s := <-sigs
			// Stop running logCollectorWorker
			/////////////////////////////////////////  ワーカーを止めること
			err := fmt.Errorf("got signal %s", s)
			slog.Error(fmt.Sprintf("%s", err))
			return
		default:
			machinesList, err := machineListReader(lc.machinesPath)
			if err != nil {
				slog.Error(fmt.Sprintf("%s", err))
			}
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
			}
			wg.Wait()
			lc.rfclient.CloseIdleConnections()
			defer cancel()
		}
	}
}
