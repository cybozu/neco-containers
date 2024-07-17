package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting pararell worker", Ordered, func() {

	// setup queue
	var m sync.Mutex
	var wg sync.WaitGroup
	var q []Machine = make([]Machine, 0)
	var lc logCollector

	BeforeAll(func() {
		ctx, cancel := context.WithCancel(context.Background())
		mq := Queue{
			queue: q,
			mu:    &m,
		}
		lc = logCollector{
			machinesPath: "testdata/configmap/serverlist.csv",
			miniNum:      1,  // 最小
			maxiNum:      10, // 最大
			currNum:      2,  // 決定コレクター数
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			ctx:          ctx, // コンテキスト
			cancel:       cancel,
			que:          mq, // コレクターのキュー
			interval:     20, // 待機秒数
			wg:           &wg,
		}
		//defer cancel()
		GinkgoWriter.Println("Start iDRAC Stub")

		bm1 := bmcMock{
			host:   "127.0.0.1:7180",
			resDir: "testdata/redfish_response_2",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm1.startMock()

		bm2 := bmcMock{
			host:   "127.0.0.1:7280",
			resDir: "testdata/redfish_response_2",
			files:  []string{"HN3CLP3-1.json", "HN3CLP3-2.json", "HN3CLP3-3.json"},
		}
		bm2.startMock()

		//startIdracMock_idrac1()
		//startIdracMock_idrac2()
		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("two workers", func() {
		var machinesList Machines
		var err error

		It("read CSV file", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("put que", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put(machinesList.machine)
		}, SpecTimeout(time.Second))

		It("start two worker", func() {
			for i := 0; i < lc.currNum; i++ {
				go lc.worker(i)
			}
			time.Sleep(20 * time.Second)
		})

		It("put que", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put(machinesList.machine)
		}, SpecTimeout(time.Second))

		It("wait two worker", func() {
			time.Sleep(20 * time.Second)
		})

	})

	AfterAll(func() {
		fmt.Println("shutdown workers")
		lc.cancel()
		time.Sleep(5 * time.Second)
	})

})
