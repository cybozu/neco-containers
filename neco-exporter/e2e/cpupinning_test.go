package e2e

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func cpuPinningInfoLines(output []byte) []string {
	var ret []string
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(line, "neco_node_cpupinning_info{") {
			ret = append(ret, strings.TrimSpace(line))
		}
	}
	return ret
}

func testCPUPinningCollector() {
	It("should stay healthy while reading the PodResources API", func() {
		Eventually(func(g Gomega) {
			output := scrapeNode(g)

			health, ok := findMetricValue(g, output, `neco_node_collector_health{collector="cpupinning"} `)
			g.Expect(ok).To(BeTrue(), "health metric not found")
			g.Expect(health).To(BeNumerically("==", 1))
		}).Should(Succeed())
	})

	It("should follow the lifecycle of a Pod with exclusive CPUs", func() {
		// podName should match testdata/cpupinning-pod.yaml
		const podName = "cpupinning-guaranteed"
		DeferCleanup(func() {
			_, _, _ = kubectl(nil, "delete", "pod", podName, "--ignore-not-found", "--wait=false")
		})

		By("checking no CPU is exclusively allocated before creating the Pod")
		Eventually(func(g Gomega) {
			g.Expect(cpuPinningInfoLines(scrapeNode(g))).To(BeEmpty())
		}).Should(Succeed())

		By("creating a Guaranteed Pod with an integer CPU request")
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "apply", "-f", "testdata/cpupinning-pod.yaml")
			kubectlSafe(g, nil, "wait", "--for=condition=Ready", "pod/"+podName, "--timeout=10s")
		}).Should(Succeed())

		By("checking cpupinning_info matches the cpuset of the container")
		Eventually(func(g Gomega) {
			cpu := strings.TrimSpace(string(kubectlSafe(g, nil, "exec", podName, "--", "cat", "/sys/fs/cgroup/cpuset.cpus.effective")))
			_, err := strconv.Atoi(cpu)
			g.Expect(err).NotTo(HaveOccurred(), "cpuset is not a single CPU: %s", cpu)

			expected := fmt.Sprintf(`neco_node_cpupinning_info{cpu="%s",node="%s",pinned_container="main",pinned_namespace="default",pinned_pod="%s"} 1`, cpu, getNodeName(g), podName)
			g.Expect(cpuPinningInfoLines(scrapeNode(g))).To(ConsistOf(expected))
		}).Should(Succeed())

		By("deleting the Pod")
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "delete", "pod", podName, "--ignore-not-found")
		}).Should(Succeed())

		By("checking cpupinning_info is removed")
		Eventually(func(g Gomega) {
			g.Expect(cpuPinningInfoLines(scrapeNode(g))).To(BeEmpty())
		}).Should(Succeed())
	})
}
