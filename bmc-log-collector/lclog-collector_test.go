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
Access the iDRAC mock and collect the lifecycle logs.
The mock serves one snapshot file per scraping cycle and returns only its
newest three entries (five for the basic scenario), as the real iDRAC
returns only the latest page (see bmcMock.redfishLclog).
*/
var _ = Describe("gathering up lifecycle logs", Ordered, func() {
	var lc logCollector
	testOutputDir := "testdata/output_lclog_collector"
	testPointerDir := "testdata/pointers_lclog_collector"

	machineBasic := Machine{Serial: "LCLOG01", BmcIP: "127.0.0.1:9180", NodeIP: "10.69.0.4"}
	machineGap := Machine{Serial: "LCLOG02", BmcIP: "127.0.0.1:9280", NodeIP: "10.69.0.5"}
	machineMismatch := Machine{Serial: "LCLOG03", BmcIP: "127.0.0.1:9380", NodeIP: "10.69.0.6"}
	machineNoLcLog := Machine{Serial: "LCLOG04", BmcIP: "127.0.0.1:9480", NodeIP: "10.69.0.7"}
	machineBadEntry := Machine{Serial: "LCLOG07", BmcIP: "127.0.0.1:9780", NodeIP: "10.69.0.10"}
	machineWriteFail := Machine{Serial: "LCLOG08", BmcIP: "127.0.0.1:9980", NodeIP: "10.69.0.12"}
	machineBadInitial := Machine{Serial: "LCLOG09", BmcIP: "127.0.0.1:10080", NodeIP: "10.69.0.13"}
	machineBadCreated := Machine{Serial: "LCLOG12", BmcIP: "127.0.0.1:10380", NodeIP: "10.69.0.16"}
	machineEmpty := Machine{Serial: "LCLOG13", BmcIP: "127.0.0.1:10480", NodeIP: "10.69.0.17"}
	machineBadNewestCreated := Machine{Serial: "LCLOG14", BmcIP: "127.0.0.1:10580", NodeIP: "10.69.0.18"}
	machines := []Machine{machineBasic, machineGap, machineMismatch, machineNoLcLog, machineBadEntry, machineWriteFail, machineBadInitial, machineBadCreated, machineEmpty, machineBadNewestCreated}

	logWriter := logTest{outputDir: testOutputDir}

	readNextLcLog := func(reader *bufio.Reader) LifecycleLog {
		GinkgoHelper()
		stringJSON, err := ReadingTestResultLogNext(reader)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("**** Received stringJSON=", stringJSON)
		var result LifecycleLog
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
				host:       machineBasic.BmcIP,
				resDir:     "testdata/redfish_response",
				lcFiles:    []string{"LCLOG01-lc-1.json", "LCLOG01-lc-2.json", "LCLOG01-lc-3.json", "LCLOG01-lc-4.json"},
				lcPageSize: 5,
			},
			{
				// More entries arrive than the page holds
				host:    machineGap.BmcIP,
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
				// An entry with a non-numeric Id appears, then the device recovers
				host:    machineBadEntry.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG07-lc-1.json", "LCLOG07-lc-2.json", "LCLOG07-lc-3.json"},
			},
			{
				// The log writer fails in the first cycle
				host:    machineWriteFail.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG08-lc-1.json", "LCLOG08-lc-2.json"},
			},
			{
				// A malformed non-first entry in the page on the initial collection
				host:    machineBadInitial.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG09-lc-1.json", "LCLOG09-lc-2.json"},
			},
			{
				// The creation time of the last read entry becomes unparsable, then the device recovers
				host:    machineBadCreated.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG12-lc-1.json", "LCLOG12-lc-2.json", "LCLOG12-lc-3.json"},
			},
			{
				// The LC log becomes empty (e.g. just cleared), then grows again
				host:    machineEmpty.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG08-lc-1.json", "LCLOG-lc-empty.json", "LCLOG03-lc-2.json"},
			},
			{
				// The creation time of the newest entry is unparsable, then the device recovers
				host:    machineBadNewestCreated.BmcIP,
				resDir:  "testdata/redfish_response",
				lcFiles: []string{"LCLOG14-lc-1.json", "LCLOG14-lc-2.json"},
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
			rfLcPath: redfishLcPath,
			ptrDir:   testPointerDir,
			username: "support",
			password: basicAuthPassword,
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

	Context("basic scenario: initial, new entries, no change, log clear", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the latest page on the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineBasic.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2", "3", "4", "5"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Serial).To(Equal(machineBasic.Serial))
				Expect(result.BmcIP).To(Equal(machineBasic.BmcIP))
				Expect(result.NodeIP).To(Equal(machineBasic.NodeIP))
				Expect(result.LogType).To(Equal("LCLog"))
			}
			Expect(lcLastReadId(machineBasic.Serial)).To(Equal(5))
		}, SpecTimeout(30*time.Second))

		It("collect only the new entries", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBasic, logWriter)

			// The page holds Id 5..9; Id 5 was read in the previous cycle
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

		It("collect the latest page after the log was cleared in iDRAC", func(ctx SpecContext) {
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

	Context("more entries arrive than the page holds", func() {
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("collect the first time", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineGap, logWriter)

			file, err = OpenTestResultLog(path.Join(testOutputDir, machineGap.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
		}, SpecTimeout(30*time.Second))

		It("emit only the entries of the latest page and skip the ones in between", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineGap, logWriter)

			// 10 entries (Id 3..12) are new, but the page holds only the newest 3
			for _, id := range []string{"10", "11", "12"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineGap.Serial)).To(Equal(12))
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

		It("collect the latest page when the same Id has a different creation time", func(ctx SpecContext) {
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

	Context("a malformed entry appears on the initial collection", func() {
		It("aborts before emitting, then collects after the device recovers", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadInitial, logWriter)
			Expect(lcLastReadId(machineBadInitial.Serial)).To(Equal(0))

			lc.collectLifecycleLog(ctx, machineBadInitial, logWriter)
			file, err := OpenTestResultLog(path.Join(testOutputDir, machineBadInitial.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader := bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2", "3"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineBadInitial.Serial)).To(Equal(3))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the creation time of the last read entry cannot be parsed", func() {
		It("aborts the cycle and keeps the pointer unchanged, then retries in the next cycle", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadCreated, logWriter)
			Expect(lcLastReadId(machineBadCreated.Serial)).To(Equal(2))

			// The page holds Id 4, 3, 2; the Created of Id 2 is not a time
			lc.collectLifecycleLog(ctx, machineBadCreated, logWriter)
			Expect(lcLastReadId(machineBadCreated.Serial)).To(Equal(2))

			// The Created of Id 2 is a time again
			lc.collectLifecycleLog(ctx, machineBadCreated, logWriter)
			file, err := OpenTestResultLog(path.Join(testOutputDir, machineBadCreated.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader := bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2", "3", "4"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineBadCreated.Serial)).To(Equal(4))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the creation time of the newest entry cannot be parsed", func() {
		It("aborts the cycle and keeps the pointer unchanged, then retries in the next cycle", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineBadNewestCreated, logWriter)
			ptr, err := readLastPointer(path.Join(testPointerDir, machineBadNewestCreated.Serial))
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LcLastReadId).To(Equal(0))
			Expect(ptr.LcLastReadCreateTime).To(Equal(int64(0)))

			// The newest entry Id 3 has a valid Created
			lc.collectLifecycleLog(ctx, machineBadNewestCreated, logWriter)
			file, err := OpenTestResultLog(path.Join(testOutputDir, machineBadNewestCreated.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader := bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2", "3"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			ptr, err = readLastPointer(path.Join(testPointerDir, machineBadNewestCreated.Serial))
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LcLastReadId).To(Equal(3))
			Expect(ptr.LcLastReadCreateTime).NotTo(Equal(int64(0)))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the LC log is empty", func() {
		It("collects nothing and keeps the pointer, then detects the clear when entries arrive", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineEmpty, logWriter)
			Expect(lcLastReadId(machineEmpty.Serial)).To(Equal(2))

			// The empty page: nothing is emitted, the position is kept, and
			// the request status is recorded as a success
			lc.collectLifecycleLog(ctx, machineEmpty, logWriter)
			ptr, err := readLastPointer(path.Join(testPointerDir, machineEmpty.Serial))
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LcLastReadId).To(Equal(2))
			Expect(ptr.LcLastHttpStatusCode).To(Equal(http.StatusOK))

			// The log grows again from Id 1 with new creation times: a clear
			lc.collectLifecycleLog(ctx, machineEmpty, logWriter)
			file, err := OpenTestResultLog(path.Join(testOutputDir, machineEmpty.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader := bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Create).To(HavePrefix("2026-09-01T00:"))
			}
			for _, id := range []string{"1", "2", "3"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
				Expect(result.Create).To(HavePrefix("2026-09-01T02:"))
			}
			Expect(lcLastReadId(machineEmpty.Serial)).To(Equal(3))
			file.Close()
		}, SpecTimeout(30*time.Second))
	})

	Context("the log writer fails", func() {
		It("does not advance the pointer, and the next cycle re-emits the entries", func(ctx SpecContext) {
			lc.collectLifecycleLog(ctx, machineWriteFail, failingLogWriter{})
			Expect(lcLastReadId(machineWriteFail.Serial)).To(Equal(0))

			lc.collectLifecycleLog(ctx, machineWriteFail, logWriter)
			file, err := OpenTestResultLog(path.Join(testOutputDir, machineWriteFail.Serial))
			Expect(err).NotTo(HaveOccurred())
			reader := bufio.NewReaderSize(file, 4096)
			for _, id := range []string{"1", "2"} {
				result := readNextLcLog(reader)
				Expect(result.Id).To(Equal(id))
			}
			Expect(lcLastReadId(machineWriteFail.Serial)).To(Equal(2))
			file.Close()
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
