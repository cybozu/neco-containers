package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testKubeletConfigCollector() {
	It("should report reserved system cpu and memory", func() {
		Eventually(func(g Gomega) {
			node := getNodeName(g)
			output := scrapeNode(g)

			cpu, ok := findMetricValue(g, output, fmt.Sprintf(`neco_node_kubelet_system_reserved{node="%s",resource="cpu",unit="core"} `, node))
			g.Expect(ok).To(BeTrue(), "cpu metric not found")
			g.Expect(cpu).To(BeNumerically("~", 0.25, 0.001))

			memory, ok := findMetricValue(g, output, fmt.Sprintf(`neco_node_kubelet_system_reserved{node="%s",resource="memory",unit="byte"} `, node))
			g.Expect(ok).To(BeTrue(), "memory metric not found")
			g.Expect(memory).To(BeNumerically("~", 512*1024*1024, 1))
		}).Should(Succeed())
	})
}
