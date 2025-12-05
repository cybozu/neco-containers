package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

const (
	kubectlPath = "../bin/kubectl"
	nsOption    = "-n=neco-exporter"
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

func kubectlGetSafe[T any](g Gomega, args ...string) *T {
	args = append([]string{"get", "-o=json"}, args...)
	stdout := kubectlSafe(g, nil, args...)

	ret := new(T)
	err := json.Unmarshal(stdout, ret)
	g.ExpectWithOffset(1, err).NotTo(HaveOccurred(), "unmarshal failed. input: %s, err: %w", err)
	return ret
}

func scrape(g Gomega, host string) []byte {
	url := fmt.Sprintf("http://%s/metrics", host)

	pilotList := kubectlGetSafe[corev1.PodList](g, "pod", "-l=app=pilot")
	g.Expect(pilotList.Items).To(HaveLen(1))
	pilotName := pilotList.Items[0].Name

	return kubectlSafe(g, nil, "exec", pilotName, "--", "curl", "-s", url)
}

func scrapeCluster(g Gomega) []byte {
	return scrape(g, "neco-cluster-exporter.neco-exporter.svc")
}

func scrapeNode(g Gomega) []byte {
	return scrape(g, "neco-node-exporter.neco-exporter.svc")
}
