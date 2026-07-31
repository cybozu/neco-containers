package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var runE2E = os.Getenv("RUN_E2E") != ""

func TestE2e(t *testing.T) {
	if !runE2E {
		t.Skip("no RUN_E2E environment variable")
	}
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(3 * time.Minute)
	SetDefaultEventuallyPollingInterval(2 * time.Second)
	SetDefaultConsistentlyDuration(15 * time.Second)
	SetDefaultConsistentlyPollingInterval(2 * time.Second)
	RunSpecs(t, "cpu-lending-agent e2e suite")
}

func setRole(role string) {
	GinkgoHelper()
	var err error
	if role == "" {
		_, err = kubectl(nil, "label", "pod", "moco-test-0", "moco.cybozu.com/role-")
	} else {
		_, err = kubectl(nil, "label", "pod", "moco-test-0", "moco.cybozu.com/role="+role, "--overwrite")
	}
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("cpu-lending-agent", Ordered, func() {
	var lent map[int]bool

	BeforeAll(func() {
		By("deploying the agent")
		_, err := kubectl(nil, "apply", "-f", "agent.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("installing the ValidatingAdmissionPolicy")
		_, err = kubectl(nil, "apply", "-f", "vap.yaml")
		Expect(err).NotTo(HaveOccurred())
		_, err = kubectl(nil, "-n", "kube-system", "rollout", "status", "ds/cpu-lending-agent", "--timeout=120s")
		Expect(err).NotTo(HaveOccurred())

		By("deploying the lender and the borrower")
		_, err = kubectl(nil, "apply", "-f", "testpods.yaml")
		Expect(err).NotTo(HaveOccurred())
		_, err = kubectl(nil, "wait", "--for=condition=Ready", "pod/moco-test-0", "--timeout=180s")
		Expect(err).NotTo(HaveOccurred())
		_, err = kubectl(nil, "wait", "--for=condition=Ready", "pod/borrower-1", "--timeout=180s")
		Expect(err).NotTo(HaveOccurred())

		lent, err = lentCPUs()
		Expect(err).NotTo(HaveOccurred())
		Expect(lent).To(HaveLen(2), "mysqld should have 2 exclusive CPUs")
	})

	It("keeps the borrower BestEffort despite the extended resource request", func() {
		out, err := kubectl(nil, "get", "pod", "borrower-1", "-o", "jsonpath={.status.qosClass}")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(Equal("BestEffort"))
	})

	It("lends the replica CPUs to the borrower", func() {
		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(containsAll(set, lent)).To(BeTrue(), "borrower cpuset %v should contain lent CPUs %v", set, lent)
		}).Should(Succeed())
	})

	It("advertises the lending capacity on the node", func() {
		Eventually(func(g Gomega) {
			capacity, allocatable, err := advertised()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(capacity).To(BeEquivalentTo(2000))
			g.Expect(allocatable).To(BeEquivalentTo(2000))
		}).Should(Succeed())
	})

	It("maintains a human-readable status annotation on the node", func() {
		Eventually(func(g Gomega) {
			out, err := kubectl(nil, "get", "node", nodeName, "-o",
				`jsonpath={.metadata.annotations.cpu-lending\.cybozu\.io/status}`)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(out)).To(ContainSubstring("capacity_milli=2000"))
			g.Expect(string(out)).To(ContainSubstring("allocated_milli=1000"))
			g.Expect(string(out)).To(ContainSubstring("free_milli=1000"))
		}).Should(Succeed())
	})

	It("reclaims the CPUs on promotion", func() {
		setRole("primary")
		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(intersects(set, lent)).To(BeFalse(), "borrower cpuset %v should not contain lent CPUs %v", set, lent)

			capacity, _, err := advertised()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(capacity).To(BeEquivalentTo(0))
		}).Should(Succeed())
	})

	It("lends again on demotion", func() {
		setRole("replica")
		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(containsAll(set, lent)).To(BeTrue())
		}).Should(Succeed())
	})

	It("records lending transitions as node events", func() {
		Eventually(func(g Gomega) {
			out, err := kubectl(nil, "get", "events", "-n", "default",
				"--field-selector", "involvedObject.kind=Node,involvedObject.uid="+nodeName+",reason=CPULendingChanged",
				"-o", "jsonpath={.items[*].message}")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(out)).To(ContainSubstring("milli-CPUs"))
		}).Should(Succeed())
	})

	It("fails closed when the role label is missing", func() {
		setRole("")
		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(intersects(set, lent)).To(BeFalse())
		}).Should(Succeed())

		setRole("replica")
		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(containsAll(set, lent)).To(BeTrue())
		}).Should(Succeed())
	})

	It("converges after an agent restart", func() {
		_, err := kubectl(nil, "-n", "kube-system", "delete", "pod", "-l", "app.kubernetes.io/name=cpu-lending-agent")
		Expect(err).NotTo(HaveOccurred())
		_, err = kubectl(nil, "-n", "kube-system", "rollout", "status", "ds/cpu-lending-agent", "--timeout=120s")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			set, err := podCPUSet("borrower-1")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(containsAll(set, lent)).To(BeTrue())
		}).Should(Succeed())
	})

	It("does not lend to non-BestEffort pods even with the request", func() {
		pod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: fake-borrower
spec:
  tolerations:
  - key: node-role.kubernetes.io/control-plane
    operator: Exists
  containers:
  - name: worker
    image: mirror.gcr.io/library/busybox
    command: ["sleep", "inf"]
    resources:
      requests:
        cpu: 100m
        %[1]s: "1000"
      limits:
        %[1]s: "1000"
`, resource)
		_, err := kubectl([]byte(pod), "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		_, err = kubectl(nil, "wait", "--for=condition=Ready", "pod/fake-borrower", "--timeout=180s")
		Expect(err).NotTo(HaveOccurred())

		Consistently(func(g Gomega) {
			set, err := podCPUSet("fake-borrower")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(intersects(set, lent)).To(BeFalse(), "Burstable pod must not receive lent CPUs")
		}).Should(Succeed())
	})

	It("rejects suspiciously small requests (the 1000m trap)", func() {
		pod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: vap-trap
spec:
  containers:
  - name: w
    image: mirror.gcr.io/library/busybox
    command: ["sleep", "inf"]
    resources:
      requests:
        %[1]s: "1000m"
      limits:
        %[1]s: "1000m"
`, resource)
		// The policy may take a moment to become active after apply.
		Eventually(func(g Gomega) {
			_, err := kubectl([]byte(pod), "apply", "-f", "-")
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("denominated in milli-CPUs"))
		}).Should(Succeed())
	})

	It("blocks placement beyond the advertised capacity", func() {
		// Stock is 2: borrower-1 and fake-borrower hold 1 each, so the next
		// borrower must stay Pending with an Insufficient message.
		pod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: borrower-overflow
spec:
  tolerations:
  - key: node-role.kubernetes.io/control-plane
    operator: Exists
  containers:
  - name: worker
    image: mirror.gcr.io/library/busybox
    command: ["sleep", "inf"]
    resources:
      requests:
        %[1]s: "1000"
      limits:
        %[1]s: "1000"
`, resource)
		_, err := kubectl([]byte(pod), "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			out, err := kubectl(nil, "get", "pod", "borrower-overflow", "-o",
				`jsonpath={.status.conditions[?(@.type=="PodScheduled")].message}`)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(out)).To(ContainSubstring("Insufficient " + resource))
		}).Should(Succeed())
	})
})
