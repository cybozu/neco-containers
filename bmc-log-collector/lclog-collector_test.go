package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

/*
Access the iDRAC mock and collect the lifecycle logs with paging.
The mock serves one snapshot file per scraping cycle and slices it into
pages of three entries (see bmcMock.redfishLclog).
*/
var _ = Describe("gathering up lifecycle logs", Ordered, func() {
	var lc logCollector
	testOutputDir := "testdata/output_lclog_collector"
	testPointerDir := "testdata/pointers_lclog_collector"

	machineBasic := Machine{Serial: "LCLOG01", BmcIP: "127.0.0.1:9180", NodeIP: "10.69.0.4"}
	machineTruncated := Machine{Serial: "LCLOG02", BmcIP: "127.0.0.1:9280", NodeIP: "10.69.0.5"}
	machineMismatch := Machine{Serial: "LCLOG03", BmcIP: "127.0.0.1:9380", NodeIP: "10.69.0.6"}
	machineNoLcLog := Machine{Serial: "LCLOG04", BmcIP: "127.0.0.1:9480", NodeIP: "10.69.0.7"}
	machineShifted := Machine{Serial: "LCLOG05", BmcIP: "127.0.0.1:9580", NodeIP: "10.69.0.8"}
	machineExhausted := Machine{Serial: "LCLOG06", BmcIP: "127.0.0.1:9680", NodeIP: "10.69.0.9"}
	machineBadEntry := Machine{Serial: "LCLOG07", BmcIP: "127.0.0.1:9780", NodeIP: "10.69.0.10"}
	machines := []Machine{machineBasic, machineTruncated, machineMismatch, machineNoLcLog, machineShifted, machineExhausted, machineBadEntry}

	logWriter := logTest{outputDir: testOutputDir}

	readNextLcLog := func(reader *bufio.Reader) LifeCycleLog {
		GinkgoHelper()
		stringJSON, err := ReadingTestResultLogNext(reader)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("**** Received stringJSON=", stringJSON)
		var result LifeCycleLog
		Expect(json.Unmarshal([]byte(stringJSON), &result)).To(Succeed())
		return result
	}

	lcLastReadId := func(serial string) int {
		GinkgoHelper()
		ptr, err := readLastPointer(path.Join(testPointerDir, serial))
		Expect(err).NotTo(HaveOccurred())
		return ptr.LcLastReadId
	}

	BeforeAll(func(ctx SpecContext) {
		for _, m := range machines {
			os.Remove(path.Join(testOutputDir, m.Serial))
			os.Remove(path.Join(testPointerDir, m.Serial))
		}
		err := os.MkdirAll(testOutputDir, 0o755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(testPointerDir, 0o755)
		Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Println("Start iDRAC Stub")
		mocks := []*bmcMock{
			{
				host:    machineBasic.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG01-lc-1.json", "LCLOG01-lc-2.json", "LCLOG01-lc-3.json", "LCLOG01-lc-4.json"},
			},
			{
				host:    machineTruncated.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG02-lc-1.json", "LCLOG02-lc-2.json"},
			},
			{
				host:    machineMismatch.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG03-lc-1.json", "LCLOG03-lc-2.json"},
			},
			{
				// The LC log service is not implemented (the mock replies 404)
				host:   machineNoLcLog.BmcIP,
				resDir: "testdata/redfish_response",
			},
			{
				// The snapshot grows by one entry between the page requests
				host:            machineShifted.BmcIP,
				resDir:          "testdata/redfish_response",
				lcFiles:         []string{"LCLOG05-lc-1.json", "LCLOG05-lc-2.json", "LCLOG05-lc-3.json", "LCLOG05-lc-4.json"},
				lcAdvanceOnSkip: true,
			},
			{
				// The log ends (no nextLink) before reaching the pointered entry
				host:    machineExhausted.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG06-lc-1.json", "LCLOG06-lc-2.json"},
			},
			{
				// An entry with a non-numeric Id appears, then the device recovers
				host:    machineBadEntry.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG07-lc-1.json", "LCLOG07-lc-2.json", "LCLOG07-lc-3.json"},
			},
		}
		for _, bm := range mocks {
			bm.accessCounter = make(map[string]int)
			bm.responseFiles = make(map[string][]string)
			bm.responseDir = make(map[string]string)
			bm.startMock()

			By("Test stub web access " + bm.host)
			Eventually(func(ctx SpecContext) error {
				req, _ := http.NewRequest("GET", "http://"+bm.host+"/", nil)
				client := &http.Client{Timeout: time.Duration(3) * time.Second}
				_, err := client.Do(req)
				return err
			}).WithContext(ctx).Should(Succeed())
		}

		lc = logCollector{
			rfLcPath:   redfishLcPath,
			ptrDir:     testPointerDir,
			username:   "support",
			password:   basicAuthPassword,
			lcMaxPages: 10,
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

	Context("basic scenario: initial, catch-up with paging, no change, log clear", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time; only the latest page is emitted", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineBasic.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"3", "4", "5"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Serial).To(Equal(machineBasic.Serial))
				Expect(result.BmcIP).To(Equal(machineBasic.BmcIP))
				Expect(result.NodeIP).To(Equal(machineBasic.NodeIP))
				Expect(result.LogType).To(Equal("LCLog"))
			}
			Expect(lcLastReadId(machineBasic.Serial)).To(Equal(5))
		}, SpecTimeout(30*time.Second))

		It("collect the new entries beyond the page boundary", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)

			for _, id := range []string{"6", "7", "8", "9"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineBasic.Serial)).To(Equal(9))
		}, SpecTimeout(30*time.Second))

		It("collect nothing when there is no new entry", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)
			Expect(lcLastReadId(machineBasic.Serial)).To(Equal(9))
			// The following test case proves that this cycle emitted nothing:
			// the next entries read from the output are the ones after the log clear.
		}, SpecTimeout(30*time.Second))

		It("restart from the latest page after the log was cleared in iDRAC", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)

			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				// The entries after the clear have new creation times
				Expect(result.Create).To(HavePrefix("2026-09-01T03:"))
			}
			Expect(lcLastReadId(machineBasic.Serial)).To(Equal(2))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("catch-up hits the page limit", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lcSmall := lc
			lcSmall.lcMaxPages = 2
			lcSmall.collectLifecycleLog(ctx, machineTruncated, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineTruncated.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
		}, SpecTimeout(30*time.Second))

		It("emit only the entries within the page limit and count up the truncated counter", func(ctx SpecContext) {
			lcSmall := lc
			lcSmall.lcMaxPages = 2
			lcSmall.collectLifecycleLog(ctx, machineTruncated, logWriter)

			// 10 entries (Id 3..12) are new, but only 2 pages x 3 entries are read
			for _, id := range []string{"7", "8", "9", "10", "11", "12"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineTruncated.Serial)).To(Equal(12))
			Expect(testutil.ToFloat64(counterLcCatchupTruncated.WithLabelValues(machineTruncated.Serial))).To(Equal(1.0))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("log clear detected by the creation time of the same Id", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineMismatch, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineMismatch.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Create).To(HavePrefix("2026-09-01T00:"))
			}
		}, SpecTimeout(30*time.Second))

		It("restart from the latest page when the same Id has a different creation time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineMismatch, logWriter)

			for _, id := range []string{"1", "2", "3"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Create).To(HavePrefix("2026-09-01T02:"))
			}
			Expect(lcLastReadId(machineMismatch.Serial)).To(Equal(3))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("an entry arrives between the page requests and shifts the $skip offset", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineShifted, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineShifted.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
		}, SpecTimeout(30*time.Second))

		It("does not emit the entries repeated by the shifted pages", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineShifted, logWriter)

			// The entry with Id 7 appears on two pages, but must be emitted once
			for _, id := range []string{"3", "4", "5", "6", "7", "8", "9"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineShifted.Serial)).To(Equal(9))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the log ends before reaching the last read entry", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineExhausted, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineExhausted.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"3", "4", "5"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
		}, SpecTimeout(30*time.Second))

		It("emit the read entries without counting the truncated counter", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineExhausted, logWriter)

			for _, id := range []string{"6", "7", "8", "9"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineExhausted.Serial)).To(Equal(9))
			Expect(testutil.ToFloat64(counterLcCatchupTruncated.WithLabelValues(machineExhausted.Serial))).To(Equal(0.0))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("an entry has a non-numeric Id", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadEntry, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineBadEntry.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineBadEntry.Serial)).To(Equal(2))
		}, SpecTimeout(30*time.Second))

		It("abort the cycle and keep the pointer unchanged", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadEntry, logWriter)
			Expect(lcLastReadId(machineBadEntry.Serial)).To(Equal(2))
			// Nothing is emitted in this cycle; the next test case proves it by
			// reading the recovered entries as the immediately following output.
		}, SpecTimeout(30*time.Second))

		It("retry successfully in the next cycle", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadEntry, logWriter)

			for _, id := range []string{"3", "4", "5"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineBadEntry.Serial)).To(Equal(5))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the BMC does not implement the LC log service", func() {
		It("does not count the 404 reply as a failure", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineNoLcLog, logWriter)
			lc.collectLifecycleLog(ctx, machineNoLcLog, logWriter)

			Expect(testutil.ToFloat64(counterRequestFailed.WithLabelValues(machineNoLcLog.Serial, metricLogTypeLc))).To(Equal(0.0))
			Expect(testutil.ToFloat64(counterRequestSuccess.WithLabelValues(machineNoLcLog.Serial, metricLogTypeLc))).To(Equal(0.0))

			ptr, err := readLastPointer(path.Join(testPointerDir, machineNoLcLog.Serial))
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LcLastHttpStatusCode).To(Equal(http.StatusNotFound))
			Expect(ptr.LcLastReadId).To(Equal(0))
			_, err = os.Stat(path.Join(testOutputDir, machineNoLcLog.Serial))
			Expect(err).To(MatchError(os.ErrNotExist))
		}, SpecTimeout(30*time.Second))
	})

	Context("pointer file compatibility", func() {
		It("read a pointer file written by an older version", func() {
			filePath := path.Join(testPointerDir, "OLDPTR1")
			err := os.WriteFile(filePath, []byte(`{"LastReadId":5,"LastError":"","LastHttpStatusCode":200,"FirstCreateTime":1234}`), 0o644)
			Expect(err).NotTo(HaveOccurred())

			ptr, err := readLastPointer(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LastReadId).To(Equal(5))
			Expect(ptr.LcLastReadId).To(Equal(0))
			Expect(ptr.LcLastReadCreateTime).To(Equal(int64(0)))
			os.Remove(filePath)
		})
	})
})
