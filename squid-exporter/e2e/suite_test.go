package e2e

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var (
	runE2E = os.Getenv("RUN_E2E") != ""
)

func TestE2e(t *testing.T) {
	if !runE2E {
		t.Skip("no RUN_E2E environment variable")
	}
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(30 * time.Second)
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	RunSpecs(t, "E2e Suite")
}

var _ = AfterSuite(func() {
	By("deleting resources from manifests")
	_, err := kubectl(nil, "delete", "-f", "./pod.yaml")
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("squid-exporter e2e test", func() {
	It("should be able to get metrics", func() {

		By("creating resources from manifests")
		_, err := kubectl(nil, "apply", "-f", "./pod.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for pod to be ready")
		Eventually(func() error {
			res, err := kubectl(nil, "get", "pod", "e2e", "-o", "json")
			if err != nil {
				return err
			}
			pod := corev1.Pod{}
			err = json.Unmarshal(res, &pod)
			if err != nil {
				return err
			}
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("pod is not ready yet")
			}
			return nil
		}).Should(Succeed())

		By("checking number of counters and metrics")
		res, err := kubectl(nil, "exec", "e2e", "-c", "squid", "--", "curl", "-s", "localhost:3128/squid-internal-mgr/counters")
		Expect(err).NotTo(HaveOccurred())
		numSquidMetrics := 0
		reader := bufio.NewScanner(bytes.NewReader(res))
		for reader.Scan() {
			line := reader.Text()
			if strings.Contains(line, "sample_time") {
				continue
			}
			numSquidMetrics++
		}

		res, err = kubectl(nil, "exec", "e2e", "-c", "squid", "--", "curl", "-s", "localhost:9100/metrics")
		Expect(err).NotTo(HaveOccurred())
		numPrometheusMetrics := 0
		reader = bufio.NewScanner(bytes.NewReader(res))
		for reader.Scan() {
			line := reader.Text()
			// exclude header line
			if strings.Contains(line, "squid_counter") {
				numPrometheusMetrics++
			}
		}
		Expect(numSquidMetrics).NotTo(BeZero())
		Expect(numPrometheusMetrics).NotTo(BeZero())
		Expect(numSquidMetrics).To(Equal(numPrometheusMetrics))

		By("checking number of service_times and metrics")
		res, err = kubectl(nil, "exec", "e2e", "-c", "squid", "--", "curl", "-s", "localhost:3128/squid-internal-mgr/service_times")
		Expect(err).NotTo(HaveOccurred())
		numSquidMetrics = 0
		reader = bufio.NewScanner(bytes.NewReader(res))
		for reader.Scan() {
			line := reader.Text()
			// exclude header line
			if strings.Contains(line, "Service Time Percentiles") {
				continue
			}
			numSquidMetrics++
		}
		res, err = kubectl(nil, "exec", "e2e", "-c", "squid", "--", "curl", "-s", "localhost:9100/metrics")
		Expect(err).NotTo(HaveOccurred())
		numPrometheusMetrics = 0
		reader = bufio.NewScanner(bytes.NewReader(res))
		for reader.Scan() {
			line := reader.Text()
			if strings.Contains(line, "squid_service_times") {
				numPrometheusMetrics++
			}
		}
		Expect(numSquidMetrics).NotTo(BeZero())
		Expect(numPrometheusMetrics).NotTo(BeZero())
		// squid's service_times has 2 metrics per line, so we need to check if counter1 * 2 == counter2
		Expect(numSquidMetrics * 2).To(Equal(numPrometheusMetrics))
	})
})
