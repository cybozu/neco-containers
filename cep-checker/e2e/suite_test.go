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

var _ = Describe("cep-checker e2e test", Ordered, func() {
	BeforeAll(func() {
		By("installing cep-checker")
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

		By("creating a curl pod from its manifest")
		_, err = kubectl(nil, "apply", "-f", "./curl.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the pod to be ready")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, "get", "pod", "curl", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			pod := corev1.Pod{}
			err = json.Unmarshal(res, &pod)
			g.Expect(err).NotTo(HaveOccurred())
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("curl pod is not ready yet")
			}
			return nil
		}).Should(Succeed())
	})

	BeforeEach(func() {
		_, err := kubectl(nil, "create", "ns", "test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_, err := kubectl(nil, "delete", "ns", "test")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should be able to get metrics", func() {
		By("creating pod from manifests")
		_, err := kubectl(nil, "apply", "-f", "./pod.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for pod to be ready")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, "get", "pod", "test", "-n", "test", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			pod := corev1.Pod{}
			err = json.Unmarshal(res, &pod)
			g.Expect(err).NotTo(HaveOccurred())
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("test pod is not ready yet")
			}
			return nil
		}).Should(Succeed())

		By("checking CEP for test pod")
		res, err := kubectl(nil, "get", "cep", "test", "-n", "test", "-o", "json")
		Expect(err).NotTo(HaveOccurred())
		cep := ciliumv2.CiliumEndpoint{}
		err = json.Unmarshal(res, &cep)
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is not found")
		res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		By("deleting test's CEP manually")
		_, err = kubectl(nil, "delete", "cep", "test", "-n", "test")
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is found")
		Eventually(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			m := strings.ReplaceAll(string(res), "\n", "")
			g.Expect(m).To(Equal(`cep_checker_missing{name="test", namespace="test", resource="cep"} 1`))
			return nil
		}).WithTimeout(time.Minute).Should(Succeed())

		By("checking metrics remains")
		Consistently(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			m := strings.ReplaceAll(string(res), "\n", "")
			g.Expect(m).To(Equal(`cep_checker_missing{name="test", namespace="test", resource="cep"} 1`))

			return nil
		}).WithTimeout(time.Minute).Should(Succeed())

		By("deleting test pod manually")
		_, err = kubectl(nil, "delete", "pod", "test", "-n", "test")
		Expect(err).NotTo(HaveOccurred())

		By("checking metrics is not found")
		Eventually(func(g Gomega) error {
			res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(res)).To(Equal(0))
			return nil
		}).WithTimeout(time.Minute).Should(Succeed())
	})

	It("should not be able to get metrics with job pods", func() {
		By("creating a job from manifests")
		_, err := kubectl(nil, "apply", "-f", "./job.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("waiting for the pod to be ready")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, "get", "pod", "-n", "test", "-l", "batch.kubernetes.io/job-name=test", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())

			pods := corev1.PodList{}
			err = json.Unmarshal(res, &pods)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pods.Items)).To(Equal(1))

			pod := pods.Items[0]
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("job pod is not ready yet")
			}
			return nil
		}).Should(Succeed())

		By("checking CEP for test pod")
		res, err := kubectl(nil, "get", "-n", "test", "cep", "-l", "job-name=test", "-o", "json")
		Expect(err).NotTo(HaveOccurred())
		ceps := ciliumv2.CiliumEndpointList{}
		err = json.Unmarshal(res, &ceps)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(ceps.Items)).To(Equal(1))

		By("checking metrics is not found")
		res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		By("deleting job pod's CEP manually")
		_, err = kubectl(nil, "delete", "-n", "test", "cep", "-l", "job-name=test")
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for the cep-checker to update the metrics")
		time.Sleep(1 * time.Minute)

		By("checking metrics is not found")
		res, err = kubectl(nil, "exec", "curl", "--", "curl", "-m", "1", "http://cep-checker-metrics.kube-system.svc:8080/metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))
	})
})
