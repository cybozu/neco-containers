package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
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
	SetDefaultEventuallyTimeout(time.Minute * 5)
	SetDefaultEventuallyPollingInterval(time.Second)
	RunSpecs(t, "E2e Suite")
}

var _ = Describe("cep-checker e2e test", func() {
	It("should be able to get metrics", func() {

		By("creating cep-checker")
		_, err := kubectl(nil, "apply", "-f", "./cep-checker.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for cep-checker to be ready")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, "get", "deploy", "-n", "kube-system", "cep-checker", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			d := appsv1.Deployment{}
			err = json.Unmarshal(res, &d)
			g.Expect(err).NotTo(HaveOccurred())

			if d.Status.Replicas != d.Status.ReadyReplicas {
				return fmt.Errorf("cep-checker is not ready")
			}

			return nil
		}).Should(Succeed())

		By("creating resources from manifests")
		_, err = kubectl(nil, "apply", "-f", "./pod.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for pod to be ready")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, "get", "pod", "test", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			pod := corev1.Pod{}
			err = json.Unmarshal(res, &pod)
			g.Expect(err).NotTo(HaveOccurred())
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("test pod is not ready yet")
			}

			res, err = kubectl(nil, "get", "pod", "curl", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			pod = corev1.Pod{}
			err = json.Unmarshal(res, &pod)
			g.Expect(err).NotTo(HaveOccurred())
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("curl pod is not ready yet")
			}
			return nil
		}).Should(Succeed())

		By("checking CEP for test pod")
		res, err := kubectl(nil, "get", "cep", "test", "-o", "json")
		Expect(err).NotTo(HaveOccurred())
		cep := ciliumv2.CiliumEndpoint{}
		err = json.Unmarshal(res, &cep)
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is not found")
		res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		By("deleting test's CEP manually")
		_, err = kubectl(nil, "delete", "cep", "test")
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is found")
		Eventually(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			m := strings.ReplaceAll(string(res), "\n", "")
			g.Expect(m).To(Equal(`cep_checker_missing{name="test", namespace="default", resource="cep"} 1`))
			return nil
		}).WithTimeout(time.Minute).Should(Succeed())

		By("checking metrics remains")
		Consistently(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			m := strings.ReplaceAll(string(res), "\n", "")
			g.Expect(m).To(Equal(`cep_checker_missing{name="test", namespace="default", resource="cep"} 1`))

			return nil
		}).WithTimeout(time.Minute).Should(Succeed())

		By("deleting test pod manually")
		_, err = kubectl(nil, "delete", "pod", "test")
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is not found")
		Eventually(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(res)).To(Equal(0))
			return nil
		}).WithTimeout(time.Minute).Should(Succeed())
	})
})
