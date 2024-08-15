package main

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

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/

var _ = Describe("gathering up logs", Ordered, func() {
	var lc selCollector
	var cl *http.Client
	var testOutputDir = "testdata/output"

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
		var machinesList []Machine
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
		lc = selCollector{
			machinesListDir: "testdata/configmap/log-collector-test.json",
			rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:          "testdata/pointers",
			username:        "user",
			password:        "pass",
			httpClient:      cl,
		}

		It("get machine list", func() {
			machinesList, err = readMachineList(lc.machinesListDir)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Machine List = ", machinesList)
		})

		// Start sel collector
		It("run logCollectorWorker", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var wg sync.WaitGroup
			logWriter := logTest{outputDir: testOutputDir}
			for _, m := range machinesList {
				wg.Add(1)
				go func() {
					lc.collectSystemEventLog(ctx, m, logWriter)
					Expect(err).NotTo(HaveOccurred())
					wg.Done()
				}()
			}
			wg.Wait()
		})

		It("verify output of collector", func(ctx SpecContext) {
			var result SystemEventLog
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial1))
			Expect(err).ToNot(HaveOccurred())

			reader = bufio.NewReaderSize(file, 4096)
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial1))
			Expect(result.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		// Start log collector (2nd)
		It("run logCollectorWorker (2nd)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			GinkgoWriter.Println("------ ", machinesList)

			// choice test logWriter to write local file
			logWriter := logTest{outputDir: testOutputDir}
			for _, m := range machinesList {
				wg.Add(1)
				go func() {
					lc.collectSystemEventLog(ctx, m, logWriter)
					Expect(err).NotTo(HaveOccurred())
					wg.Done()
				}()
			}
			defer cancel()
			wg.Wait()
		})

		It("verify output of collector (2nd)", func(ctx SpecContext) {
			var result SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial1))
			Expect(result.Id).To(Equal("2"))

			file.Close()
		}, SpecTimeout(3*time.Second))
	})
})
