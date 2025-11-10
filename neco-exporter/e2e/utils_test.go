package e2e

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega"
)

const (
	kubectlPath = "../bin/kubectl"
)

func runCommand(path string, input []byte, args ...string) ([]byte, []byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(path, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	if err := cmd.Run(); err != nil {
		_, file := filepath.Split(path)
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("%s failed with %s: stderr=%s", file, err, stderr)
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

func kubectl(input []byte, args ...string) ([]byte, []byte, error) {
	return runCommand(kubectlPath, input, args...)
}

func kubectlSafe(g Gomega, input []byte, args ...string) []byte {
	stdout, stderr, err := kubectl(input, args...)
	g.Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
	return stdout
}

func scrape(g Gomega, svc string) []byte {
	const nsOption = "-n=neco-exporter"

	stdout := kubectlSafe(g, nil, "get", "service", nsOption, svc, "-o=jsonpath={ .spec.ports[0].nodePort }")
	nodePort := string(stdout)

	url := fmt.Sprintf("http://localhost:%s/metrics", nodePort)
	stdout, stderr, err := runCommand("docker", nil, "exec", "neco-exporter-control-plane", "curl", "-s", url)
	g.ExpectWithOffset(2, err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s, err: %v", string(stdout), string(stderr), err)

	return stdout
}

func scrapeCluster(g Gomega) []byte {
	return scrape(g, "neco-cluster-exporter")
}

func scrapeServer(g Gomega) []byte {
	return scrape(g, "neco-server-exporter")
}
