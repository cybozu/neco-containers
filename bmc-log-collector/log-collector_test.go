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
	var lc logCollector
	var cl *http.Client
	testOutputDir := "testdata/output_log_collector"
	testPointerDir := "testdata/pointers_log_collector"
	serial := "683FPQ3"
	metricsPath := "/testmetrics2"
	metricsPort := ":29000"

	// Start iDRAC Stub
	BeforeAll(func(ctx SpecContext) {
		os.Remove(path.Join(testOutputDir, serial))
		os.Remove(path.Join(testPointerDir, serial))
		err := os.MkdirAll(testOutputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(testPointerDir, 0o755)
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
		lc = logCollector{
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
An entry with a non-numeric Id appears in the SEL, then the device recovers.
The collector aborts the cycle without updating the pointer file and
retries in the next cycle (the same policy as the LC log collector).
*/
var _ = Describe("SEL entry with a non-numeric Id", Ordered, func() {
	var lc logCollector
	testOutputDir := "testdata/output_log_collector"
	testPointerDir := "testdata/pointers_log_collector"
	machine := Machine{Serial: "SELBAD01", BmcIP: "127.0.0.1:9880", NodeIP: "10.69.0.11"}
	logWriter := logTest{outputDir: testOutputDir}

	var file *os.File
	var reader *bufio.Reader

	readNextSel := func() SystemEventLog {
		GinkgoHelper()
		stringJSON, err := ReadingTestResultLogNext(reader)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("**** Received stringJSON=", stringJSON)
		var result SystemEventLog
		Expect(json.Unmarshal([]byte(stringJSON), &result)).To(Succeed())
		return result
	}

	selLastReadId := func() int {
		GinkgoHelper()
		ptr, err := readLastPointer(path.Join(testPointerDir, machine.Serial))
		Expect(err).NotTo(HaveOccurred())
		return ptr.LastReadId
	}

	BeforeAll(func(ctx SpecContext) {
		os.Remove(path.Join(testOutputDir, machine.Serial))
		os.Remove(path.Join(testPointerDir, machine.Serial))
		err := os.MkdirAll(testOutputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(testPointerDir, 0o755)
		Expect(err).NotTo(HaveOccurred())

		bm := bmcMock{
			host:          machine.BmcIP,
			resDir:        "testdata/redfish_response",
			files:         []string{"SELBAD01-1.json", "SELBAD01-2.json", "SELBAD01-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm.startMock()

		By("Test stub web access " + bm.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		lc = logCollector{
			rfSelPath: redfishPath,
			ptrDir:    testPointerDir,
			username:  "support",
			password:  basicAuthPassword,
			httpClient: &http.Client{
				Timeout: time.Duration(10) * time.Second,
				Transport: &http.Transport{
					TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
					DisableKeepAlives:   true,
					TLSHandshakeTimeout: 20 * time.Second,
					DialContext: (&net.Dialer{
						Timeout: 15 * time.Second,
					}).DialContext,
				},
			},
		}
	}, NodeTimeout(30*time.Second))

	It("collect the first time", func(ctx SpecContext) {
		lc.collectSystemEventLog(ctx, machine, logWriter)

		var err error
		file, err = OpenTestResultLog(path.Join(testOutputDir, machine.Serial))
		Expect(err).NotTo(HaveOccurred())
		reader = bufio.NewReaderSize(file, 4096)
		for _, id := range []string{"1", "2"} {
			Expect(readNextSel().Id).To(Equal(id))
		}
		Expect(selLastReadId()).To(Equal(2))
	}, SpecTimeout(30*time.Second))

	It("abort the cycle and keep the pointer unchanged", func(ctx SpecContext) {
		lc.collectSystemEventLog(ctx, machine, logWriter)
		// All the Ids are validated before writing any entry, so nothing is
		// emitted and the pointer stays; the next test case proves it by
		// reading the recovered entries as the immediately following output.
		Expect(selLastReadId()).To(Equal(2))
	}, SpecTimeout(30*time.Second))

	It("retry successfully in the next cycle", func(ctx SpecContext) {
		lc.collectSystemEventLog(ctx, machine, logWriter)

		for _, id := range []string{"3", "4"} {
			Expect(readNextSel().Id).To(Equal(id))
		}
		Expect(selLastReadId()).To(Equal(4))
		file.Close()
	}, SpecTimeout(30*time.Second))
})
