package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	var lc logCollector
	var tr *http.Transport
	var cl *http.Client
	var mu sync.Mutex

	// Start iDRAC Stub
	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cl = &http.Client{
			Timeout:   time.Duration(10) * time.Second,
			Transport: tr,
		}

		lc = logCollector{
			machinesPath: "testdata/configmap/log-collector-test.json",
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			testMode:     true,
			testOut:      "testdata/output",
			user:         "user",
			password:     "pass",
			rfclient:     cl,
			mutex:        &mu,
		}
		GinkgoWriter.Println("Start iDRAC Stub")
		bm := bmcMock{
			host:   "127.0.0.1:8180",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}

		bm.startMock()
		time.Sleep(10 * time.Second)
	})

	Context("single worker with go-routine", func() {
		var machinesList Machines
		var err error
		var serial1 string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader

		It("get machine list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Machine List = ", machinesList)
		})

		// ワーカースレッドの停止のテストが欲しい

		// Start log collector
		It("run worker with the go routine (1st time)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(1 * time.Second)
			}
			wg.Wait()
			defer cancel()
		})

		It("verify output of collector (1st time)", func() {
			var result SystemEventLog
			file, err = OpenTestResultLog(path.Join(lc.testOut, serial1))
			Expect(err).ToNot(HaveOccurred())

			// Read test log
			reader = bufio.NewReaderSize(file, 4096)

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial1))
			Expect(result.Id).To(Equal("1"))
		})

		// Start log collector
		It("run worker with the go routine (2nd time)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			GinkgoWriter.Println("------ ", machinesList.Machine)
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(1 * time.Second)
			}
			defer cancel()
			wg.Wait()
		})

		It("verify output of collector (2nd time)", func() {
			var result SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial1))
			Expect(result.Id).To(Equal("2"))
		})
		file.Close()
	})
	AfterAll(func() {
		fmt.Println("shutdown workers")
		cl.CloseIdleConnections()
		time.Sleep(5 * time.Second)
	})
})
