package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testNICIRQCollector() {
	// kind nodes have no TxRx queue interrupts; only the health of the walk can be checked.
	It("should stay healthy while walking /proc/irq", func() {
		Eventually(func(g Gomega) {
			output := scrapeNode(g)

			health, ok := findMetricValue(g, output, `neco_node_collector_health{collector="nicirq"} `)
			g.Expect(ok).To(BeTrue(), "health metric not found")
			g.Expect(health).To(BeNumerically("==", 1))
		}).Should(Succeed())
	})
}
