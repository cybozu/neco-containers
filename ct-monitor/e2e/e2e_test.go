package e2e

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

const (
	kubeResourcePath = "./testdata/install.yaml"
)

var _ = Describe("ct-monitor e2e test", func() {
	BeforeEach(func() {
		By("applying resources")
		_, err := kubectl(nil, "apply", "-f", kubeResourcePath)
		Expect(err).NotTo(HaveOccurred())

		By("suspending cronjob")
		_, err = kubectl(nil, "patch", "cronjob", "ct-monitor", "-p", `{"spec": {"suspend": true}}`)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("deleting resources")
		_, _ = kubectl(nil, "delete", "--ignore-not-found", "-f", kubeResourcePath)
	})

	It("should run cronjob", func() {
		By("triggering job from cronjob")
		_, err := kubectl(nil, "create", "job", "--from=cronjob/ct-monitor", "ct-monitor-test-1")
		Expect(err).NotTo(HaveOccurred())

		ctMonitorListArgs := []string{
			"get", "pods",
			"--field-selector=status.phase==Succeeded",
			"-l", "app.kubernetes.io/name=ct-monitor",
			"-o", "json",
		}

		By("waiting for ct-monitor job pod to succeed")
		Eventually(func(g Gomega) error {
			res, err := kubectl(nil, ctMonitorListArgs...)
			g.Expect(err).NotTo(HaveOccurred())

			podList := corev1.PodList{}
			err = json.Unmarshal(res, &podList)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(podList.Items)).To(Equal(1))
			return nil
		}).Should(Succeed())

		By("checking ct-monitor logs")
		res, err := kubectl(nil, "logs", "-l", "app.kubernetes.io/name=ct-monitor")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(res)).NotTo(ContainSubstring("error"))
		Expect(string(res)).To(ContainSubstring("done checking"))
	})
})
