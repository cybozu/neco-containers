package e2e

import (
	"encoding/json"
	"os"
	"testing"
	"time"

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

var _ = Describe("daemonset-updater e2e test", func() {
	It("should update pods correctly", func() {
		By("checking the daemonset desiring image version")
		Eventually(func(g Gomega) error {
			out, err := kubectl(nil, "get", "ds", "testhttpd", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			ds := appsv1.DaemonSet{}
			err = json.Unmarshal(out, &ds)
			g.Expect(err).NotTo(HaveOccurred())
			image := ds.Spec.Template.Spec.Containers[0].Image
			g.Expect(image).To(Equal("ghcr.io/cybozu/testhttpd:0.2.5"))

			return nil
		}).Should(Succeed())

		By("checking that pods owned by the daemonset don't have the desired image version")
		Eventually(func(g Gomega) error {
			out, err := kubectl(nil, "get", "pod", "-l", "app=testhttpd", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			podList := corev1.PodList{}
			err = json.Unmarshal(out, &podList)
			g.Expect(err).NotTo(HaveOccurred())
			for _, pod := range podList.Items {
				g.Expect(pod.Spec.Containers[0].Image).To(Equal("ghcr.io/cybozu/testhttpd:0.2.4"))
			}

			return nil
		}).Should(Succeed())

		By("creating a daemonset-updater as a job")
		_, err := kubectl(nil, "apply", "-f", "job.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("checking the job pod is completed")
		Eventually(func(g Gomega) error {
			out, err := kubectl(nil, "get", "pod", "-l", "batch.kubernetes.io/job-name=daemonset-updater", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			podList := corev1.PodList{}
			err = json.Unmarshal(out, &podList)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(podList.Items)).To(Equal(1))
			g.Expect(string(podList.Items[0].Status.Phase)).To(Equal("Succeeded"))

			return nil
		}).Should(Succeed())

		By("checking pods are updated except a node(daemonset-updater-worker)")
		Eventually(func(g Gomega) error {
			out, err := kubectl(nil, "get", "pod", "-l", "app=testhttpd", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			podList := corev1.PodList{}
			err = json.Unmarshal(out, &podList)
			g.Expect(err).NotTo(HaveOccurred())
			for _, pod := range podList.Items {
				if pod.Spec.NodeName == "daemonset-updater-worker" {
					g.Expect(pod.Spec.Containers[0].Image).To(Equal("ghcr.io/cybozu/testhttpd:0.2.4"))
				} else {
					g.Expect(pod.Spec.Containers[0].Image).To(Equal("ghcr.io/cybozu/testhttpd:0.2.5"))
				}
			}

			return nil
		}).Should(Succeed())

		By("checking all nodes are schedulable")
		Eventually(func(g Gomega) error {
			out, err := kubectl(nil, "get", "node", "-o", "json")
			g.Expect(err).NotTo(HaveOccurred())
			nodeList := corev1.NodeList{}
			err = json.Unmarshal(out, &nodeList)
			g.Expect(err).NotTo(HaveOccurred())
			for _, node := range nodeList.Items {
				for _, taint := range node.Spec.Taints {
					g.Expect(taint.Key).NotTo(Equal("node.kubernetes.io/unschedulable"))
				}
			}
			return nil
		}).Should(Succeed())
	})
})
