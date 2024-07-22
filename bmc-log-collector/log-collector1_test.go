package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	var wg sync.WaitGroup
	var lc logCollector
	var mq MessageQueue

	// Start iDRAC Stub
	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")
		ctx, cancel := context.WithCancel(context.Background())
		mq.queue = make(chan Machine, 1000)

		lc = logCollector{
			machinesPath: "testdata/configmap/log-collector-test.csv",
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
			testMode:     true,
			testOut:      "testdata/output/",
		}
		GinkgoWriter.Println("Start iDRAC Stub")
		bm := bmcMock{
			host:   "127.0.0.1:9080",
			resDir: "testdata/redfish_response",
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
		var rslt, rslt2 SystemEventLog
		var serial1 string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader

		It("get machine list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		// ブロックする可能性があるので追加が必要
		It("put que for worker", func(ctx SpecContext) {
			fmt.Println(machinesList.machine)
			lc.que.put3(machinesList.machine)
		}, SpecTimeout(time.Second))

		// Start log collector
		It("get SEL by bmcClient", func() {
			v := lc.que.get2()
			byteData, err := bmcClient("https://" + v.BmcIP + lc.rfUrl)
			GinkgoWriter.Println("got log =", string(byteData))
			Expect(err).NotTo(HaveOccurred())
		})

		It("put machine list to queue again for test", func() {
			GinkgoWriter.Println("Put que ==", machinesList.machine)
			lc.que.put3(machinesList.machine)
			l := lc.que.len2()
			Expect(l).To(Equal(1))
		})

		It("Check output SEL 1st", func() {
			go lc.worker(1)
			time.Sleep(3 * time.Second)
			for {
				file, err = os.Open(path.Join(lc.testOut, serial1))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				reader = bufio.NewReaderSize(file, 4096)
				stringJSON, _ := reader.ReadString('\n')
				json.Unmarshal([]byte(stringJSON), &rslt)
				GinkgoWriter.Println("------ ", string(rslt.Serial))
				GinkgoWriter.Println("------ ", string(rslt.Id))
				break
			}
			Expect(rslt.Serial).To(Equal(serial1))
			Expect(rslt.Id).To(Equal("1"))
		})

		It("Check output SEL 2nd", func() {
			stringJSON, _ := reader.ReadString('\n')
			fmt.Println("*3 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial1))
			Expect(rslt2.Id).To(Equal("2"))
		})
		file.Close()

	})
	AfterAll(func() {
		fmt.Println("shutdown workers")
		lc.cancel()
		time.Sleep(5 * time.Second)
	})
})
