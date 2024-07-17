package main

import (
	"context"
	"fmt"
	"github.com/cybozu-go/log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// start main-loop at main()
func main() {
	// setup queue
	var m sync.Mutex
	var wg sync.WaitGroup
	var q []Machine = make([]Machine, 0)

	// setup context
	ctx, cancel := context.WithCancel(context.Background())

	mq := Queue{
		queue: q,
		mu:    &m,
	}
	lc := logCollector{
		machinesPath: "testdata/conf/serverlist.csv",
		miniNum:      1,  // 最小
		maxiNum:      10, // 最大
		currNum:      0,  // 決定コレクター数
		rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:       "pointers",
		ctx:          ctx, // コンテキスト
		que:          mq,  // コレクターのキュー
		interval:     20,  // 待機秒数
		wg:           &wg,
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// check parameter
	numCollector, _ := strconv.Atoi(os.Getenv("LOG_COLLECTOR"))

	if numCollector < lc.miniNum {
		slog.Error("less than minimum_collectors")
		os.Exit(1)
	} else if numCollector > lc.miniNum {
		slog.Error("greater than the maximum number")
		os.Exit(1)
	}
	lc.currNum = numCollector

	// start log collectors
	for i := 0; i < lc.currNum; i++ {
		go lc.worker(i)
	}
	defer cancel()

	// Main loop
	for {
		select {
		case <-sigs:
			s := <-sigs
			slog.Error(fmt.Sprintf("Got signal %s", s))
			return
		default:
			machineList, err := machineListReader(lc.machinesPath)
			if err != nil {
				log.Error("cat not read server list", nil)
			}
			lc.que.put(machineList.machine)
		}
		// wait until next collecting cycle
		time.Sleep(lc.interval * time.Second)
	}
}
