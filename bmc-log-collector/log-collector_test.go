package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
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
Read the machines list and access iDRAC mock, and eliminate duplicated entry.
*/
var _ = Describe("gathering up logs", Ordered, func() {
	var lc selCollector
	var cl *http.Client
	var testOutputDir = "testdata/output_log_collector"
	var testPointerDir = "testdata/pointers_log_collector"
	var serial = "683FPQ3"
	var metricsPath = "/testmetrics2"
	var metricsPort = ":29000"

	// Start iDRAC Stub
	BeforeAll(func(ctx SpecContext) {
		os.Remove(path.Join(testOutputDir, serial))
		os.Remove(path.Join(testPointerDir, serial))
		err := os.MkdirAll(testOutputDir, 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(testPointerDir, 0755)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Start iDRAC Stub")
		bm1 := bmcMock{
			host:          "127.0.0.1:8180",
			resDir:        "testdata/redfish_response",
			files:         []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm1.startMock()

		// Wait starting stub servers
		By("Test stub web access" + bm1.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm1.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		// Must start metrics exporter, if not it will get SIGSEGV
		go func() {
			metrics(metricsPath, metricsPort)
		}()
	}, NodeTimeout(10*time.Second))

	Context("SEL collector test", func() {
		var machinesList []Machine
		var err error
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
			ptrDir:          testPointerDir,
			username:        "support",
			password:        basicAuthPassword,
			httpClient:      cl,
		}

		It("get machine list", func() {
			machinesList, err = readMachineList(lc.machinesListDir)
			Expect(err).NotTo(HaveOccurred())
			GinkgoWriter.Println("Machine List = ", machinesList)
		})

		It("collect iDRAC log (run1)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var wg sync.WaitGroup

			// Choice the test logWriter to write a local file
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

		It("verify output (run1)", func(ctx SpecContext) {
			var result SystemEventLog
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())

			reader = bufio.NewReaderSize(file, 4096)
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial))
			Expect(result.Id).To(Equal("1"))
		}, SpecTimeout(10*time.Second))

		It("collect iDRAC log (run2)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			GinkgoWriter.Println("------ ", machinesList)

			// Choice the test logWriter to write a local file
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

		It("verify output (run2)", func(ctx SpecContext) {
			var result SystemEventLog
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial))
			Expect(result.Id).To(Equal("2"))

		}, SpecTimeout(10*time.Second))

		It("collect iDRAC log (run3)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup

			// Choice the test logWriter to write local file
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

		It("verify output (run3)", func(ctx SpecContext) {
			var result SystemEventLog
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &result)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(result.Serial))
			GinkgoWriter.Println("-------- id = ", string(result.Id))
			Expect(result.Serial).To(Equal(serial))
			Expect(result.Id).To(Equal("3"))
			file.Close()
		}, SpecTimeout(10*time.Second))
	})
})
