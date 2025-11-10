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
			scrapeServer(g)
		}).Should(Succeed())
	})

	It("should report collectors health", func() {
		// mock collector is designed to fail returning metrices sometimes.
		var healthOK, healthNG, duration bool
		Eventually(func(g Gomega) {
			output := string(scrapeCluster(g))

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
}
