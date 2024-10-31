package main

/*
  Read the machines list and access iDRAC mock, and eliminate duplicated entry.
*/
import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {
	var testOutputDir = "testdata/output_main_test"
	var testPointerDir = "testdata/pointers_main_test"
	var serial1 = "683FPQ3" // basic test case1
	var serial2 = "HN3CLP3" // basic test case2
	var serial3 = "J7N6MW3" // the log reset test case
	var serial4 = "483FPQ3" // the "OEM software events" with random timestamp test case
	var serial5 = "JQ5M8K3" // Duplicate event occur case

	BeforeAll(func(ctx SpecContext) {
		GinkgoWriter.Println("start BMC stub servers")
		os.Remove(path.Join(testOutputDir, serial1))
		os.Remove(path.Join(testPointerDir, serial1))
		os.Remove(path.Join(testOutputDir, serial2))
		os.Remove(path.Join(testPointerDir, serial2))
		os.Remove(path.Join(testOutputDir, serial3))
		os.Remove(path.Join(testPointerDir, serial3))
		os.Remove(path.Join(testOutputDir, serial4))
		os.Remove(path.Join(testPointerDir, serial4))
		os.Remove(path.Join(testOutputDir, serial5))
		os.Remove(path.Join(testPointerDir, serial5))

		err := os.MkdirAll(testPointerDir, 0750)
		Expect(err).ToNot(HaveOccurred())
		err = os.MkdirAll(testOutputDir, 0750)
		Expect(err).ToNot(HaveOccurred())

		bm1 := bmcMock{
			host:          "127.0.0.1:7180",
			resDir:        "testdata/redfish_response",
			files:         []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm1.startMock()

		bm2 := bmcMock{
			host:          "127.0.0.1:7280",
			resDir:        "testdata/redfish_response",
			files:         []string{"HN3CLP3-1.json", "HN3CLP3-2.json", "HN3CLP3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm2.startMock()

		bm3 := bmcMock{
			host:          "127.0.0.1:7380",
			resDir:        "testdata/redfish_response",
			files:         []string{"J7N6MW3-1.json", "J7N6MW3-2.json", "J7N6MW3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm3.startMock()

		bm4 := bmcMock{
			host:          "127.0.0.1:7480",
			resDir:        "testdata/redfish_response",
			files:         []string{"483FPQ3-1.json", "483FPQ3-2.json", "483FPQ3-3.json", "483FPQ3-4.json", "483FPQ3-5.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm4.startMock()

		bm5 := bmcMock{
			host:          "127.0.0.1:7580",
			resDir:        "testdata/redfish_response",
			files:         []string{"JQ5M8K3-1.json", "JQ5M8K3-2.json", "JQ5M8K3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm5.startMock()

		// Wait starting stub servers
		By("Test stub web access" + bm1.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm1.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		By("Test stub web access" + bm2.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm2.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		By("Test stub web access" + bm3.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm3.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		By("Test stub web access" + bm4.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm4.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

		By("Test stub web access" + bm5.host)
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", "http://"+bm5.host+"/", nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())

	}, NodeTimeout(10*time.Second))

	Context("stub of main equivalent", func() {
		It("main loop test", func() {
			intervalTimeString := "10s"
			intervalTime, _ := time.ParseDuration(intervalTimeString)
			lcConfig := selCollector{
				machinesListDir: "testdata/configmap/serverlist2.json",
				rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
				ptrDir:          testPointerDir,
				username:        "support",
				intervalTime:    intervalTime,
			}
			user, err := LoadBMCUserConfig("testdata/etc/bmc-user.json")
			Expect(err).ToNot(HaveOccurred())
			lcConfig.password = user.Support.Password.Raw

			// Setup logWriter for test
			logWriter := logTest{
				outputDir: testOutputDir,
			}
			func() {
				go doLogScrapingLoop(lcConfig, logWriter)
			}()
			// Stop scraper after 30 sec
			time.Sleep(30 * time.Second)
		})
	})

	Context("verify 683FPQ3", func() {
		var serial string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("1st reply", func(ctx SpecContext) {
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(3*time.Second))

		It("2nd reply", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))
	})

	Context("verify HN3CLP3", func() {
		var serial string = "HN3CLP3"
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("check 1st log record", func(ctx SpecContext) {
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(39*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))

		It("check 3rd log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(30*time.Second))

		It("check 4th log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("4"))
		}, SpecTimeout(30*time.Second))
	})

	Context("verify J7N6MW3", func() {
		var serial string = "J7N6MW3"
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("check 1st log record", func(ctx SpecContext) {
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(30*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))

		It("check 3rd log record which after SEL cleanup", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(30*time.Second))

		It("check 4th log record which after SEL cleanup", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))

		It("check 5th log record which after SEL cleanup", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(30*time.Second))
	})

	Context("verify 483FPQ3", func() {
		var serial string = "483FPQ3"
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("check 1st log record", func(ctx SpecContext) {
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(30*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))

		It("check 3rd log record with OEM software event", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(30*time.Second))

		It("check 4th log record with OEM software event", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("4"))
		}, SpecTimeout(30*time.Second))

		It("check 5th log record that is null", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).To(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)
			Expect(len(stringJSON)).To(Equal(0))
		}, SpecTimeout(30*time.Second))
	})

	Context("verify JQ5M8K3", func() {
		var serial string = "JQ5M8K3"
		var file *os.File
		var reader *bufio.Reader
		var err error

		It("check 1st log record", func(ctx SpecContext) {
			file, err = OpenTestResultLog(path.Join(testOutputDir, serial))
			Expect(err).ToNot(HaveOccurred())
			reader = bufio.NewReaderSize(file, 4096)

			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("1"))
		}, SpecTimeout(30*time.Second))

		It("check 2nd log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("2"))
		}, SpecTimeout(30*time.Second))

		It("check 3rd log record with OEM software event", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("3"))
		}, SpecTimeout(30*time.Second))

		It("check 4th log record with OEM software event", func(ctx SpecContext) {
			for i := 4; i <= 12; i++ {
				stringJSON, err := ReadingTestResultLogNext(reader)
				Expect(err).ToNot(HaveOccurred())
				GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

				var rslt SystemEventLog
				err = json.Unmarshal([]byte(stringJSON), &rslt)
				Expect(err).ToNot(HaveOccurred())
				GinkgoWriter.Println("------ ", string(rslt.Serial))
				GinkgoWriter.Println("------ ", string(rslt.Message))
				Expect(rslt.Serial).To(Equal(serial))
				Expect(rslt.Id).To(Equal(strconv.Itoa(i)))
				if i == 12 {
					Expect(rslt.Message).To(Equal("The system inlet temperature is greater than the upper warning threshold."))
				}
			}
		}, SpecTimeout(30*time.Second))

		It("check 5th log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			GinkgoWriter.Println("------ ", string(rslt.Message))
			Expect(rslt.Message).To(Equal("The system inlet temperature is within range."))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("13"))
		}, SpecTimeout(30*time.Second))

		It("check 6th log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			GinkgoWriter.Println("------ ", string(rslt.Message))
			Expect(rslt.Message).To(Equal("The system inlet temperature is greater than the upper warning threshold."))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("14"))
		}, SpecTimeout(30*time.Second))

		It("check 7th log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			GinkgoWriter.Println("------ ", string(rslt.Message))
			Expect(rslt.Message).To(Equal("The system inlet temperature is within range."))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("15"))
		}, SpecTimeout(30*time.Second))

		It("check 8th log record", func(ctx SpecContext) {
			stringJSON, err := ReadingTestResultLogNext(reader)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			var rslt SystemEventLog
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			GinkgoWriter.Println("------ ", string(rslt.Message))
			Expect(rslt.Message).To(Equal("The system inlet temperature is greater than the upper warning threshold."))
			Expect(rslt.Serial).To(Equal(serial))
			Expect(rslt.Id).To(Equal("16"))
		}, SpecTimeout(30*time.Second))

		It("check suppressing log", func(ctx SpecContext) {
			_, err := ReadingTestResultLogNext(reader)
			Expect(err).To(HaveOccurred())
		}, SpecTimeout(30*time.Second))
	})
})
