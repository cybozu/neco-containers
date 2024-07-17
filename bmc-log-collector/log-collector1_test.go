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

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	// setup queue
	var m sync.Mutex
	var wg sync.WaitGroup
	var q []Machine = make([]Machine, 0)
	var lc logCollector

	// Start iDRAC Stub
	BeforeAll(func() {
		ctx, cancel := context.WithCancel(context.Background())
		mq := Queue{
			queue: q,
			mu:    &m,
		}
		lc = logCollector{
			machinesPath: "testdata/configmap/bmc-list-it.csv",
			miniNum:      1,  // 最小
			maxiNum:      10, // 最大
			currNum:      0,  // 決定コレクター数
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			ctx:          ctx, // コンテキスト
			cancel:       cancel,
			que:          mq, // コレクターのキュー
			interval:     20, // 待機秒数
			wg:           &wg,
		}
		GinkgoWriter.Println("Start iDRAC Stub")
		bm := bmcMock{
			host:   "127.0.0.1:9080",
			resDir: "testdata/redfish_response_1",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm.startMock()
		time.Sleep(10 * time.Second)
	})

	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("single worker", func() {
		var machinesList Machines
		var err error
		It("get machine list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		// ブロックする可能性があるので追加が必要
		It("put que for worker", func(ctx SpecContext) {
			fmt.Println(machinesList.machine)
			lc.que.put(machinesList.machine)
		}, SpecTimeout(time.Second))

		// Start log collector
		It("get SEL by bmcClient", func() {
			v := lc.que.get()
			byteData, err := bmcClient("https://" + v.BmcIP + lc.rfUrl)
			GinkgoWriter.Println("got log =", string(byteData))
			Expect(err).NotTo(HaveOccurred())
		})

		It("put machine list to queue again for test", func() {
			GinkgoWriter.Println("Put que ==", machinesList.machine)
			lc.que.put(machinesList.machine)
			l := lc.que.len()
			Expect(l).To(Equal(1))
		})

		It("output SEL", func() {
			go lc.worker(1)
			time.Sleep(30 * time.Second)
			// テスト確認方法？
		})
	})
	AfterAll(func() {
		fmt.Println("shutdown workers")
		lc.cancel()
		time.Sleep(5 * time.Second)
	})
})
