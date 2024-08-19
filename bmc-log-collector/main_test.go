package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/
import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {
	BeforeAll(func() {
		GinkgoWriter.Println("start BMC stub servers")
		os.Remove("testdata/pointers")
		os.Remove("testdata/output")
		os.MkdirAll("testdata/pointers", 0750)
		os.MkdirAll("testdata/output", 0750)

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
			lcConfig := selCollector{
				machinesListDir: "testdata/configmap/serverlist2.json",
				rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
				ptrDir:          "testdata/pointers",
				username:        "user",
				password:        "pass",
				intervalTime:    10, // sec
			}

			// setup logWriter for test
			logWriter := logTest{
				outputDir: "testdata/output",
			}
			func() {
				go doLogScrapingLoop(lcConfig, logWriter)
			}()
			// stop scraper after 15 sec
			time.Sleep(60 * time.Second)
		})
	})

	Context("verify metrics", func() {
		var metricsLines []string
		It("get metrics", func() {
			url := "http://localhost:8080/metrics"
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{Timeout: time.Duration(10) * time.Second}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			buf, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(string(buf))
			defer resp.Body.Close()
			metricsLines = strings.Split(string(buf), "\n")
			fmt.Println(metricsLines)
		})

		It("verify HELP line in metrics", func() {
			Expect(metricsLines[0]).To(Equal("# HELP failed_counter The failed count for Redfish of BMC accessing"))
		})
		It("verify TYPE line in metrics", func() {
			Expect(metricsLines[1]).To(Equal("# TYPE failed_counter counter"))
		})

		It("iDRAC 683FPQ3 127.0.0.1:7180", func() {
			metricsLine := metricsLines[2]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("failed_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7180"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("683FPQ3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("404"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("2", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("iDRAC HN3CLP3 127.0.0.1:7280", func() {
			metricsLine := metricsLines[3]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("failed_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7280"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("HN3CLP3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("404"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("2", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("iDRAC J7N6MW3 127.0.0.1:7380", func() {
			metricsLine := metricsLines[4]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("failed_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7380"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("J7N6MW3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("404"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("2", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("verify HELP line in metrics", func() {
			Expect(metricsLines[5]).To(Equal("# HELP success_counter The success count for Redfish of BMC accessing"))
		})

		It("verify TYPE line in metrics", func() {
			Expect(metricsLines[6]).To(Equal("# TYPE success_counter counter"))
		})

		It("iDRAC 683FPQ3 127.0.0.1:7180", func() {
			metricsLine := metricsLines[7]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("success_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7180"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("683FPQ3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("200"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("3", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("iDRAC HN3CLP3 127.0.0.1:7280", func() {
			metricsLine := metricsLines[8]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("success_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7280"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("HN3CLP3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("200"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("3", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("iDRAC J7N6MW3 127.0.0.1:7380", func() {
			metricsLine := metricsLines[9]
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine + "\n"))
			if err != nil {
				fmt.Println("err ", err)
			}

			for _, v := range metricsFamily {
				GinkgoWriter.Printf("name=%s, type=%s \n", v.GetName(), v.GetType())
				Expect(v.GetName()).To(Equal("success_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					switch idx {
					case 0:
						Expect(l.GetName()).To(Equal("ip_addr"))
						Expect(l.GetValue()).To(Equal("127.0.0.1:7380"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("J7N6MW3"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("200"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("3", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))
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
		}, SpecTimeout(39*time.Second))

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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))
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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))

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
		}, SpecTimeout(30*time.Second))
	})

})
