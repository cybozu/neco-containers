package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testExporter() {
	It("should scrape", func() {
		Eventually(func(g Gomega) {
			scrapeCluster(g)
			scrapeClusterLeader(g)
			scrapeClusterNonLeader(g)
			scrapeNode(g)
		}).Should(Succeed())
	})

	It("should report leader", func() {
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))
			g.Expect(output).To(ContainSubstring(`neco_cluster_collector_leader 1`))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := string(scrapeClusterNonLeader(g))
			g.Expect(output).To(ContainSubstring(`neco_cluster_collector_leader 0`))
		}).Should(Succeed())
	})

	It("should report collectors health", func() {
		// mock collector is designed to fail returning metrics sometimes.
		var healthOK, healthNG, duration bool
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))

			if strings.Contains(output, `neco_cluster_collector_health{collector="mock"} 1`) {
				healthOK = true
			}
			if strings.Contains(output, `neco_cluster_collector_health{collector="mock"} 0`) {
				healthNG = true
			}
			if strings.Contains(output, `neco_cluster_collector_process_seconds{collector="mock"}`) {
				duration = true
			}

			g.Expect(healthOK && healthNG && duration).To(BeTrue())
		}).Should(Succeed())
	})

	It("should not report leader-collectors health from non-leader", func() {
		Consistently(func(g Gomega) {
			output := string(scrapeClusterNonLeader(g))

			healthOK := strings.Contains(output, `neco_cluster_collector_health{collector="mock"} 1`)
			healthNG := strings.Contains(output, `neco_cluster_collector_health{collector="mock"} 0`)
			duration := strings.Contains(output, `neco_cluster_collector_process_seconds{collector="mock"}`)

			g.Expect(healthOK).NotTo(BeTrue())
			g.Expect(healthNG).NotTo(BeTrue())
			g.Expect(duration).NotTo(BeTrue())
		}).Should(Succeed())
	})

	It("should report leader metrics", func() {
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))
			g.Expect(output).To(ContainSubstring("neco_cluster_mock_test"))
		}).Should(Succeed())

		Consistently(func(g Gomega) {
			output := string(scrapeClusterNonLeader(g))
			g.Expect(output).NotTo(ContainSubstring("neco_cluster_mock_test"))
		}).Should(Succeed())
	})
}
