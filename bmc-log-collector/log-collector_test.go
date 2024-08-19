package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
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

		// must start metrics exporter, if not it get SIGSEGV
		go func() {
			metrics()
		}()
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
			})

			It("verify HELP line in metrics", func() {
				Expect(metricsLines[0]).To(Equal("# HELP success_counter The success count for Redfish of BMC accessing"))
			})

			It("verify TYPE line in metrics", func() {
				Expect(metricsLines[1]).To(Equal("# TYPE success_counter counter"))
			})

			It("iDRAC 683FPQ3 127.0.0.1:8180", func() {
				metricsLine := metricsLines[2]
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
							Expect(l.GetValue()).To(Equal("127.0.0.1:8180"))
						case 1:
							Expect(l.GetName()).To(Equal("serial"))
							Expect(l.GetValue()).To(Equal("683FPQ3"))
						case 2:
							Expect(l.GetName()).To(Equal("status"))
							Expect(l.GetValue()).To(Equal("200"))
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
		})
	})
})
