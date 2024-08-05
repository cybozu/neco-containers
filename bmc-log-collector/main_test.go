package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		os.Remove("testdata/pointers/HN3CLP3")
		os.Remove("testdata/output/HN3CLP3")

		os.Remove("testdata/pointers/J7N6MW3")
		os.Remove("testdata/output/J7N6MW3")

		GinkgoWriter.Println("start BMC stub servers")

		bm1 := bmcMock{
			host:   "127.0.0.1:7180",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm1.startMock()

		bm2 := bmcMock{
			host:   "127.0.0.1:7280",
			resDir: "testdata/redfish_response",
			files:  []string{"HN3CLP3-1.json", "HN3CLP3-2.json", "HN3CLP3-3.json"},
		}
		bm2.startMock()

		bm3 := bmcMock{
			host:   "127.0.0.1:7380",
			resDir: "testdata/redfish_response",
			files:  []string{"J7N6MW3-1.json", "J7N6MW3-2.json", "J7N6MW3-3.json"},
		}
		bm3.startMock()

		// Wait starting stub servers
		time.Sleep(10 * time.Second)

	})

	Context("stub of main equivalent", func() {
		It("main loop test", func() {
			os.Setenv("BMC_USERNAME", "user")
			os.Setenv("BMC_PASSWORD", "pass")
			lc := selCollector{
				machinesListDir: "testdata/configmap/serverlist2.json",
				rfUriSel:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
				ptrDir:          "testdata/pointers",
				username:        "user",
				password:        "pass",
			}

			// setup logWriter for test
			logWriter := logTest{outputDir: "testdata/output"}
			doLogScrapingLoop(lc, logWriter)
		})
	})

	Context("verify 683FPQ3", func() {
		var serial string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader
		var err error
		var testOut string = "testdata/output"

		It("1st reply", func(ctx SpecContext) {
			var rslt SystemEventLog
			file, err = OpenTestResultLog(path.Join(testOut, serial))
			Expect(err).ToNot(HaveOccurred())

			reader = bufio.NewReaderSize(file, 4096)
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		It("2nd reply", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(3*time.Second))
	})

	Context("verify HN3CLP3", func() {
		var serial string = "HN3CLP3"
		var file *os.File
		var reader *bufio.Reader
		var err error
		var testOut string = "testdata/output"

		It("check 1st log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			file, err = OpenTestResultLog(path.Join(testOut, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(3*time.Second))

		It("check 3rd log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(3*time.Second))

		It("check 4th log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("4"))
		}, SpecTimeout(3*time.Second))
	})

	Context("verify J7N6MW3", func() {
		var serial string = "J7N6MW3"
		var file *os.File
		var reader *bufio.Reader
		var err error
		var testOut string = "testdata/output"

		It("check 1st log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			file, err = OpenTestResultLog(path.Join(testOut, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(3*time.Second))

		It("check 3rd log record which after SEL cleanup", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		It("check 4th log record which after SEL cleanup", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(3*time.Second))

		It("check 5th log record which after SEL cleanup", func(ctx SpecContext) {
			var rslt SystemEventLog

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(3*time.Second))

	})

	AfterAll(func() {
		fmt.Println("shutdown stub servers")
	})
})
