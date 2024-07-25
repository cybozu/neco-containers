package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	//"strconv"
	"sync"
	"syscall"
	//"time"
)

// start main-loop at main()
func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	//var mq MessageQueue
	//mq.queue = make(chan Machine, 1000)

	lc := logCollector{
		machinesPath: "testdata/conf/serverlist.csv",
		//miniNum:      1,  // 最小
		//maxiNum:      10, // 最大
		//currNum:      0,  // 決定コレクター数
		rfUrl:  "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir: "pointers",
		//ctx:          ctx, // コンテキスト
		//que:          mq,  // コレクターのキュー
		//interval: 20, // 待機秒数
		wg: &wg,
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	/*
		// check parameter
		numCollector, err := strconv.Atoi(os.Getenv("LOG_COLLECTOR"))
		if err != nil {
			slog.Error(fmt.Sprintf("%s", err))
			return
		}
	*/

	/*
		if numCollector < lc.miniNum {
			err := fmt.Errorf("less than minimum_collectors")
			slog.Error(fmt.Sprintf("%s", err))
			os.Exit(1)
		} else if numCollector > lc.miniNum {
			err := fmt.Errorf("greater than the maximum number")
			slog.Error(fmt.Sprintf("%s", err))
			os.Exit(1)
		}
		lc.currNum = numCollector
	*/

	// start log collectors
	/*
		for i := 0; i < lc.currNum; i++ {
			go lc.worker(i)
		}
		defer cancel()
	*/

	// Main loop
	for {
		select {
		case <-sigs:
			s := <-sigs
			err := fmt.Errorf("got signal %s", s)
			slog.Error(fmt.Sprintf("%s", err))
			return
		default:
			machinesList, err := machineListReader(lc.machinesPath)
			if err != nil {
				slog.Error(fmt.Sprintf("%s", err))
			}
			//lc.que.put3(machineList.machine)
			for i := 0; i < len(machinesList.Machine); i++ {
				//q.queue <- m[i]
				// セットアップ
				//x := machineList.machine[i].BmcIP
				//y := machineList.machine[i].Serial
				//z := machineList.machine[i].NodeIP
				go lc.worker(ctx, machinesList.Machine[i])
			}
			defer cancel()
		}
		// wait until next collecting cycle
		//time.Sleep(lc.interval * time.Second)
	}
}
