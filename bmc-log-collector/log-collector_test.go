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

// lookbackForAllEntries is long enough to output all the entries in testdata on the first read.
const lookbackForAllEntries = 100 * 365 * 24 * time.Hour

/*
Read the machines list and access iDRAC mock, and eliminate duplicated entry.
*/
var _ = Describe("gathering up logs", Ordered, func() {
	var lc selCollector
	var cl *http.Client
	testOutputDir := "testdata/output_log_collector"
	serial := "683FPQ3"
	metricsPath := "/testmetrics2"
	metricsPort := ":29000"

	// Start iDRAC Stub
	BeforeAll(func(ctx SpecContext) {
		os.Remove(path.Join(testOutputDir, serial))
		err := os.MkdirAll(testOutputDir, 0o755)
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
			ptrStore:        newPointerStore(),
			startupLookback: lookbackForAllEntries,
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
				wg.Go(func() {
					lc.collectSystemEventLog(ctx, m, logWriter)
					Expect(err).NotTo(HaveOccurred())
				})
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
				wg.Go(func() {
					lc.collectSystemEventLog(ctx, m, logWriter)
					Expect(err).NotTo(HaveOccurred())
				})
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
				wg.Go(func() {
					lc.collectSystemEventLog(ctx, m, logWriter)
					Expect(err).NotTo(HaveOccurred())
				})
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

/*
On the first read after start, only the entries created within the startup lookback are output.
*/
var _ = Describe("startup lookback of the collector", Ordered, func() {
	testOutputDir := "testdata/output_startup_lookback"
	serial := "683FPQ3"
	host := "127.0.0.1:8280"
	var lc selCollector
	machine := Machine{Serial: serial, BmcIP: host, NodeIP: "10.1.0.1"}
	logWriter := logTest{outputDir: testOutputDir}

	BeforeAll(func(ctx SpecContext) {
		os.Remove(path.Join(testOutputDir, serial))
		err := os.MkdirAll(testOutputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())

		bm := bmcMock{
			host:          host,
			resDir:        "testdata/redfish_response",
			files:         []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm.startMock()
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+host+"/", nil)
			client := &http.Client{Timeout: 3 * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		lc = selCollector{
			rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrStore:        newPointerStore(),
			startupLookback: time.Hour,
			username:        "support",
			password:        basicAuthPassword,
			httpClient: &http.Client{
				Timeout:   10 * time.Second,
				Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			},
		}
	}, NodeTimeout(10*time.Second))

	It("does not output old entries on the first read", func(ctx SpecContext) {
		lc.collectSystemEventLog(ctx, machine, logWriter)
		Expect(lc.ptrStore.get(serial)).To(HaveField("LastReadId", 1))
		_, err := os.Stat(path.Join(testOutputDir, serial))
		Expect(os.IsNotExist(err)).To(BeTrue())
	}, SpecTimeout(10*time.Second))

	It("outputs new entries after the first read", func(ctx SpecContext) {
		lc.collectSystemEventLog(ctx, machine, logWriter)
		lc.collectSystemEventLog(ctx, machine, logWriter)

		file, err := OpenTestResultLog(path.Join(testOutputDir, serial))
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()
		reader := bufio.NewReaderSize(file, 4096)
		for _, id := range []string{"2", "3"} {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).NotTo(HaveOccurred())
			var result SystemEventLog
			Expect(json.Unmarshal([]byte(stringJSON), &result)).To(Succeed())
			Expect(result.Id).To(Equal(id))
		}
	}, SpecTimeout(20*time.Second))
})
