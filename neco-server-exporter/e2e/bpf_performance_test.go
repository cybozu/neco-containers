package e2e

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testBPFPerformance() {
	It("should scrape", func() {
		const nsOption = "-n=neco-server-exporter"

		stdout := kubectlSafe(Default, nil, "get", "service", nsOption, "neco-server-exporter", "-o=jsonpath={ .spec.ports[0].nodePort }")
		nodePort := string(stdout)

		url := fmt.Sprintf("http://localhost:%s/metrics", nodePort)
		stdout, stderr, err := runCommand("docker", nil, "exec", "neco-server-exporter-control-plane", "curl", "-s", url)
		Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", string(stdout), string(stderr), err)

		reader := bufio.NewScanner(bytes.NewReader(stdout))
		foundRunTimeSeconds := false
		foundRunCountTotal := false
		for reader.Scan() {
			line := reader.Text()
			switch {
			case strings.Contains(line, "neco_server_bpf_run_time_seconds_total"):
				if strings.Contains(line, "cil_from_container") {
					foundRunTimeSeconds = true
				}
			case strings.Contains(line, "neco_server_bpf_run_count_total"):
				if strings.Contains(line, "cil_from_container") {
					foundRunCountTotal = true
				}
			}
		}
		Expect(foundRunTimeSeconds).To(BeTrue())
		Expect(foundRunCountTotal).To(BeTrue())
	})
}
