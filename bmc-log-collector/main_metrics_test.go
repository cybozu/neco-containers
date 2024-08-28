package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/
import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {
	var testOutputDir = "testdata/output_main_test"
	var testPointerDir = "testdata/pointers_main_test"
	var serial1 = "683FPQ3"
	var serial2 = "HN3CLP3"
	var serial3 = "J7N6MW3"
	var metricsPath = "/metrics"
	var metricsPort = ":8080"
	var wg sync.WaitGroup

	cl := &http.Client{
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
	lcConfig := selCollector{
		machinesListDir: "testdata/configmap/serverlist2.json", // これを途中で差し替える必要がある
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrDir:          testPointerDir,
		username:        "user",
		password:        "pass",
		httpClient:      cl,
	}

	BeforeAll(func() {
		GinkgoWriter.Println("start BMC stub servers")
		os.Remove(path.Join(testOutputDir, serial1))
		os.Remove(path.Join(testPointerDir, serial1))
		os.Remove(path.Join(testOutputDir, serial2))
		os.Remove(path.Join(testPointerDir, serial2))
		os.Remove(path.Join(testOutputDir, serial3))
		os.Remove(path.Join(testPointerDir, serial3))

		os.MkdirAll(testPointerDir, 0750)
		os.MkdirAll(testOutputDir, 0750)

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

		// Start metrics server
		go metrics("/metrics", ":8080")

		// Wait starting stub servers
		time.Sleep(10 * time.Second)

	})

	Context("1st scrape cycle", func() {
		var machines []Machine
		var err error

		It("get machine list", func() {
			machines, err = readMachineList(lcConfig.machinesListDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(machines)).To(Equal(3))
		})

		It("scraping logs from iDRAC", func(ctx SpecContext) {
			ctx0 := context.Background()
			// setup logWriter for test
			logWriter := logTest{
				outputDir: testOutputDir,
			}
			for _, m := range machines {
				fmt.Println("m", m)
				wg.Add(1)
				go func() {
					lcConfig.collectSystemEventLog(ctx0, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
		}, SpecTimeout(30*time.Second))

		It("drop metrics of machine that is retired", func(ctx SpecContext) {
			err = dropMetricsWhichRetiredMachine(lcConfig.ptrDir, machines)
			Expect(err).ToNot(HaveOccurred())
		})

		// メトリックスの取得
		It("get metrics", func() {
			var metricsLines []string

			url := "http://localhost" + metricsPort + metricsPath
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{Timeout: time.Duration(10) * time.Second}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			buf, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			metricsLines = strings.Split(string(buf), "\n")
			for i, v := range metricsLines {
				GinkgoWriter.Println(i, v)
			}
			GinkgoWriter.Println(metricsLines)
		})

	})

	Context("2nd scrape cycle", func() {
		var machines []Machine
		var err error
		var metricsLines []string

		It("get machine list", func() {
			// ２回目
			lcConfig.machinesListDir = "testdata/configmap/serverlist3.json"
			machines, err = readMachineList(lcConfig.machinesListDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(machines)).To(Equal(2))
		})

		It("collect logs from iDRAC", func(ctx SpecContext) {
			// setup logWriter for test
			logWriter := logTest{
				outputDir: testOutputDir,
			}
			for _, m := range machines {
				wg.Add(1)
				go func() {
					lcConfig.collectSystemEventLog(ctx, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
		}, SpecTimeout(3*time.Second))

		It("collect logs from iDRAC", func(ctx SpecContext) {
			err = dropMetricsWhichRetiredMachine(lcConfig.ptrDir, machines)
			Expect(err).ToNot(HaveOccurred())
		})

		// メトリックスの取得
		It("get metrics", func() {
			url := "http://localhost" + metricsPort + metricsPath
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{Timeout: time.Duration(10) * time.Second}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			buf, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			metricsLines = strings.Split(string(buf), "\n")
			for i, v := range metricsLines {
				GinkgoWriter.Println(i, v)
			}
			GinkgoWriter.Println(metricsLines)
		})

		// metricsLines を１行づつ調べて、退役したレコードに相当するものが無いことを確認
		It("verify drop iDRAC", func() {
			metricsLine, err := findMetrics(metricsLines, "success_counter")
			Expect(err).NotTo(HaveOccurred())
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine))
			if err != nil {
				GinkgoWriter.Println("err ", err)
			}
			for _, v := range metricsFamily {
				GinkgoWriter.Printf("xxxxxxxxxxxxxxxxxx name=%s, type=%s \n", v.GetName(), v.GetType())
				//	Expect(v.GetName()).To(Equal("success_counter"))
				//l := len(v.GetMetric())
				for i, v := range v.GetMetric() {
					fmt.Println(" ================= i", i, "v", v)
				}

			}
			for _, v := range metricsFamily {
				fmt.Println("xxxxxxxxxxxx v.GetMetric()", v.GetMetric())
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("xxxxxxxxxxxxxxxxxxxxxxx idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
					/*
						if l.GetValue() == "172.16.0.1" {
							switch idx {
							case 0:
								Expect(l.GetName()).To(Equal("ip_addr"))
								Expect(l.GetValue()).To(Equal("172.16.0.1"))
							case 1:
								Expect(l.GetName()).To(Equal("serial"))
								Expect(l.GetValue()).To(Equal("ABC123X"))
							}
							GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
							f, err := strconv.ParseFloat("1", 64)
							if err != nil {
								GinkgoWriter.Printf("error %w", err)
							}
							Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
						}
					*/
				}
			}
		})

	})

	Context("3rd scrape cycle", func() {
		var machines []Machine
		var err error

		It("get machine list", func() {
			// ３回目
			lcConfig.machinesListDir = "testdata/configmap/serverlist3.json"
			machines, err = readMachineList(lcConfig.machinesListDir)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("machines list", machines)
			Expect(len(machines)).To(Equal(2))
		})

		It("collect logs from iDRAC", func(ctx SpecContext) {
			// setup logWriter for test
			logWriter := logTest{
				outputDir: testOutputDir,
			}
			for _, m := range machines {
				wg.Add(1)
				go func() {
					lcConfig.collectSystemEventLog(ctx, m, logWriter)
					wg.Done()
				}()
			}
			wg.Wait()
		}, SpecTimeout(3*time.Second))

		It("collect logs from iDRAC", func(ctx SpecContext) {
			err = dropMetricsWhichRetiredMachine(lcConfig.ptrDir, machines)
			Expect(err).ToNot(HaveOccurred())
		})

		// メトリックスの取得
		It("get metrics", func() {
			var metricsLines []string

			url := "http://localhost" + metricsPort + metricsPath
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{Timeout: time.Duration(10) * time.Second}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			buf, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			metricsLines = strings.Split(string(buf), "\n")
			for i, v := range metricsLines {
				GinkgoWriter.Println(i, v)
			}
			GinkgoWriter.Println(metricsLines)
		})

	})

})
