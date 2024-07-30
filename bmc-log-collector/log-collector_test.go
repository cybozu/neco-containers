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
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gathering up logs", Ordered, func() {

	var lc logCollector
	var cl *http.Client
	var mu sync.Mutex

	// Start iDRAC Stub
	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		GinkgoWriter.Println("Start iDRAC Stub")
		bm := bmcMock{
			host:   "127.0.0.1:8180",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm.startMock()
		time.Sleep(10 * time.Second)
	})

	Context("log collector function test", func() {
		var machinesList Machines
		var err error
		var serial1 string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader

		cl = &http.Client{
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
		lc = logCollector{
			machinesPath: "testdata/configmap/log-collector-test.json",
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			testMode:     true,
			testOut:      "testdata/output",
			user:         "user",
			password:     "pass",
			rfClient:     cl,
			mutex:        &mu,
		}

		It("get machine list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Machine List = ", machinesList)
		})

		// Start log collector
		It("run logCollectorWorker", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var wg sync.WaitGroup
			for i := 0; i < len(machinesList.Machine); i++ {
				wg.Add(1)
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
				time.Sleep(1 * time.Second)
			}
			wg.Wait()
		})

		It("verify output of collector", func(ctx SpecContext) {
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
		}, SpecTimeout(3*time.Second))

		// Start log collector (2nd)
		It("run logCollectorWorker (2nd)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			GinkgoWriter.Println("------ ", machinesList.Machine)
			for i := 0; i < len(machinesList.Machine); i++ {
				wg.Add(1)
				go lc.logCollectorWorker(ctx, &wg, machinesList.Machine[i])
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(1 * time.Second)
			}
			defer cancel()
			wg.Wait()
		})

		It("verify output of collector (2nd)", func(ctx SpecContext) {
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

			file.Close()
		}, SpecTimeout(3*time.Second))

	})
	AfterAll(func() {
		fmt.Println("shutdown workers")
		cl.CloseIdleConnections()
	})
})
