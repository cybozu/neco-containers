package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
)

var _ = Describe("Get Metrics export", Ordered, func() {
	var metricsPath = "/testmetrics1"
	var metricsPort = ":28000"
	BeforeAll(func() {
		go func() {
			metrics(metricsPath, metricsPort)
		}()
	})

	Context("Normal", func() {
		var metricsLines []string
		It("put metrics at failed case", func() {
			counterRequestFailed.WithLabelValues("404", "ABC123X", "172.16.0.1").Inc()
		})
		It("get metrics at success case", func() {
			counterRequestSuccess.WithLabelValues("200", "ABC123X", "172.16.0.1").Inc()
		})
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
			fmt.Println(metricsLines)
		})

		It("verify HELP line in metrics", func() {
			Expect(searchMetricsComment(metricsLines, "# HELP failed_counter The failed count for Redfish of BMC accessing")).To(Equal(true))
		})
		It("verify TYPE line in metrics", func() {
			Expect(searchMetricsComment(metricsLines, "# TYPE failed_counter counter")).To(Equal(true))
		})

		It("iDRAC ABC123X 172.16.0.1 failed", func() {
			metricsLine, err := findMetrics(metricsLines, "failed_counter")
			Expect(err).NotTo(HaveOccurred())

			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine))
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
					if l.GetValue() == "172.16.0.1" {
						switch idx {
						case 0:
							Expect(l.GetName()).To(Equal("ip_addr"))
							Expect(l.GetValue()).To(Equal("172.16.0.1"))
						case 1:
							Expect(l.GetName()).To(Equal("serial"))
							Expect(l.GetValue()).To(Equal("ABC123X"))
						case 2:
							Expect(l.GetName()).To(Equal("status"))
							Expect(l.GetValue()).To(Equal("404"))
						}
						GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
						f, err := strconv.ParseFloat("1", 64)
						if err != nil {
							GinkgoWriter.Printf("error %w", err)
						}
						Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
					}
				}
			}
		})

		It("verify HELP line in metrics", func() {
			Expect(searchMetricsComment(metricsLines, "# HELP success_counter The success count for Redfish of BMC accessing")).To(Equal(true))
		})

		It("verify TYPE line in metrics", func() {
			Expect(searchMetricsComment(metricsLines, "# TYPE success_counter counter")).To(Equal(true))
		})

		It("iDRAC ABC123X 172.16.0.1 success", func() {
			metricsLine, err := findMetrics(metricsLines, "success_counter")
			Expect(err).NotTo(HaveOccurred())
			p := expfmt.TextParser{}
			metricsFamily, err := p.TextToMetricFamilies(strings.NewReader(metricsLine))
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
					if l.GetValue() == "172.16.0.1" {
						switch idx {
						case 0:
							Expect(l.GetName()).To(Equal("ip_addr"))
							Expect(l.GetValue()).To(Equal("172.16.0.1"))
						case 1:
							Expect(l.GetName()).To(Equal("serial"))
							Expect(l.GetValue()).To(Equal("ABC123X"))
						case 2:
							Expect(l.GetName()).To(Equal("status"))
							Expect(l.GetValue()).To(Equal("200"))
						}
						GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
						f, err := strconv.ParseFloat("1", 64)
						if err != nil {
							GinkgoWriter.Printf("error %w", err)
						}
						Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
					}
				}
			}
		})
	})
})
