package main

import (
	//"fmt"
	//"io"
	//"net/http"
	//"time"

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

/*
Test metrics
テストでＢＭＣのモックを起動して、スクレイプを実行、
スクレイプの結果がメトリックスとして正しいことを検証する
メトリックスの定義とコレクターの書き方を調べること。
*/
var _ = Describe("Get Metrics export", Ordered, func() {
	BeforeAll(func() {
		go func() {
			metrics()
		}()
	})

	Context("Normal", func() {
		var metricsLines []string
		It("get metrics", func() {
			counterRequestFailed.WithLabelValues("404", "ABC123X", "172.16.0.1").Inc()
		})
		It("get metrics", func() {
			counterRequestSuccess.WithLabelValues("200", "ABC123X", "172.16.0.1").Inc()
		})
		It("get metrics", func() {
			url := "http://localhost:8080/metrics"
			req, err := http.NewRequest("GET", url, nil)
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{Timeout: time.Duration(10) * time.Second}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			buf, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(string(buf), "==========================================")
			metricsLines = strings.Split(string(buf), "\n")
			fmt.Println(metricsLines)
		})

		It("verify HELP line in metrics", func() {
			Expect(metricsLines[0]).To(Equal("# HELP failed_counter The failed count for Redfish of BMC accessing"))
		})
		It("verify TYPE line in metrics", func() {
			Expect(metricsLines[1]).To(Equal("# TYPE failed_counter counter"))
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
				Expect(v.GetName()).To(Equal("failed_counter"))
			}

			for _, v := range metricsFamily {
				for idx, l := range v.GetMetric()[0].Label {
					GinkgoWriter.Printf("idx=%d  label name=%s, value=%s \n", idx, l.GetName(), l.GetValue())
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
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("1", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})

		It("verify HELP line in metrics", func() {
			Expect(metricsLines[3]).To(Equal("# HELP success_counter The success count for Redfish of BMC accessing"))
		})

		It("verify TYPE line in metrics", func() {
			Expect(metricsLines[4]).To(Equal("# TYPE success_counter counter"))
		})

		It("iDRAC 683FPQ3 127.0.0.1:8180", func() {
			metricsLine := metricsLines[5]
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
						Expect(l.GetValue()).To(Equal("172.16.0.1"))
					case 1:
						Expect(l.GetName()).To(Equal("serial"))
						Expect(l.GetValue()).To(Equal("ABC123X"))
					case 2:
						Expect(l.GetName()).To(Equal("status"))
						Expect(l.GetValue()).To(Equal("200"))
					}
				}
				GinkgoWriter.Printf("untyped value=%f \n", v.GetMetric()[0].Untyped.GetValue())
				f, err := strconv.ParseFloat("1", 64)
				if err != nil {
					GinkgoWriter.Printf("error %w", err)
				}
				Expect(v.GetMetric()[0].Untyped.GetValue()).To(Equal(f))
			}
		})
	})
})
