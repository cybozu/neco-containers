package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/cybozu-go/log"
)

// queue and mutex
var queue []Machine
var wg sync.WaitGroup
var mu sync.Mutex

// Get queue
func getQueue() Machine {
	var v Machine
	for {
		fmt.Println("read que")
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
func collector(n int, url string) {
	wg.Add(1)
	defer wg.Done()
	for {
		v := getQueue()
		url := "https://" + v.BmcIP + url
		b, err := bmcClient(url)
		if err != nil {
			log.Error("Error", nil)
		}
		printLogs(b, v)
		if n == 0 {
			return
		}
	}
}

func main() {

	var redfish_url = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	queue = make([]Machine, 0)

	// Start iDRAC log collector
	wn := 3 // 起動数
	for i := 0; i < wn; i++ {
		go collector(i, redfish_url) // ログコレクター起動
	}

	// Main loop
	for {
		machineList, err := machineListReader("conf/serverlist.csv")
		if err != nil {
			log.Error("cat not read server list", nil)
		}
		putQueue(machineList.machine)
		time.Sleep(20 * time.Second)
	}
}
