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

// queue and mutex
var queue []Machine
var wg sync.WaitGroup
var mu sync.Mutex
var global_wg sync.WaitGroup

// Get queue
func getQueue() Machine {
	var v Machine
	for {
		slog.Debug("read que")
		if len(queue) == 0 {
			time.Sleep(1 * time.Second)
		} else {
			mu.Lock()
			v = queue[0]
			queue = queue[1:]
			mu.Unlock()
			break
		}
	}
	return v
}

// Put queue
func putQueue(m []Machine) {
	mu.Lock()
	queue = m
	mu.Unlock()
}

// Log collector
func collector(ctx context.Context, n int, url string, ptr string) {
	fmt.Printf("collector #%d Started\n", n)
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			//slog.Error("log-collectors is canceled ")
			fmt.Printf("log-collectors stopped\n")
			return
		default:
			target := getQueue()
			slog.Debug(fmt.Sprintf("Worker %d", n))
			url := "https://" + target.BmcIP + url
			byteBuf, err := bmcClient(url)
			if err != nil {
				log.Error("Error", nil)
			}
			printLogs(byteBuf, target, ptr)
		}
	}
}

func runMainLoop(global_ctx context.Context) {
	// static parameter
	var minimum_collectors = 1
	var maximum_collectors = 10
	var num_collector = 0
	var redfish_url = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	var ptrDir = "pointers"

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// queue is in target machines
	queue = make([]Machine, 0)

	// start iDRAC log collector
	ctx := context.Background()
	ctxParent, cancel := context.WithCancel(ctx)

	// check parameter
	num_collector, _ = strconv.Atoi(os.Getenv("LOG_COLLECTOR"))
	if num_collector < minimum_collectors {
		slog.Error("less than minimum_collectors")
		num_collector = minimum_collectors
	} else if num_collector > maximum_collectors {
		slog.Error("greater than the maximum number")
		num_collector = maximum_collectors
	}

	// start log collectors
	for i := 0; i < num_collector; i++ {
		go collector(ctxParent, i, redfish_url, ptrDir)
	}
	defer cancel()

	// Main loop
	for {
		select {
		case <-sigs:
			s := <-sigs
			fmt.Println("Got signal:", s)
			cancel()
			return
		case <-global_ctx.Done():
			fmt.Println("log-collectors is canceled ")
			cancel()
			//global_wg.Done()
			return
		default:
			machineList, err := machineListReader("conf/serverlist.csv")
			if err != nil {
				log.Error("cat not read server list", nil)
			}
			putQueue(machineList.machine)
		}
		// wait until next collecting cycle
		time.Sleep(20 * time.Second)
	}
}

// start main-loop at main()
func main() {
	global_ctx, global_cancel := context.WithCancel(context.Background())
	global_wg.Add(1)
	go runMainLoop(global_ctx)
	defer global_cancel()
	global_wg.Wait()
	time.Sleep(20 * time.Second)
}
