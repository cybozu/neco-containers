package e2e

import (
	"bufio"
	"bytes"
	"errors"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testBPFCollector() {
	It("should report necessary metrices", func() {
		remaining := []string{
			"neco_server_bpf_run_time_seconds_total",
			"neco_server_bpf_run_count_total",
		}

		Eventually(func(g Gomega) {
			stdout := scrape(g)
			reader := bufio.NewScanner(bytes.NewReader(stdout))
			for reader.Scan() {
				line := reader.Text()
				remaining = slices.DeleteFunc(remaining, func(s string) bool {
					return strings.Contains(line, s)
				})
			}
			g.Expect(remaining).To(BeEmpty())
		}).Should(Succeed())
	})

	It("should report long program names using BTF", func() {
		Eventually(func(g Gomega) error {
			stdout := scrape(g)
			reader := bufio.NewScanner(bytes.NewReader(stdout))
			for reader.Scan() {
				line := reader.Text()
				if strings.Contains(line, "neco_server_bpf_run_time_seconds_total") &&
					strings.Contains(line, `name="cil_from_container"`) {
					return nil
				}
			}
			return errors.New("failed to find long program name")
		}).Should(Succeed())
	})
}
