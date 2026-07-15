package e2e

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// findMetricValue scans Prometheus exposition text for a line starting with prefix
// (a metric name plus its label set) and returns the trailing sample value.
func findMetricValue(g Gomega, output []byte, prefix string) (float64, bool) {
	reader := bufio.NewScanner(bytes.NewReader(output))
	for reader.Scan() {
		line := reader.Text()
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		v, err := strconv.ParseFloat(value, 64)
		g.Expect(err).NotTo(HaveOccurred(), "failed to parse metric value from line: %s", line)
		return v, true
	}
	return 0, false
}

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
