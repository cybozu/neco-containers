package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"
	coordinationv1 "k8s.io/api/coordination/v1"
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

func pilotSafe(g Gomega, input []byte, args ...string) []byte {
	args = append([]string{"exec", "-i", "deploy/pilot", "--"}, args...)
	return kubectlSafe(g, input, args...)
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

func scrapeClusterLeader(g Gomega) []byte {
	lease := kubectlGetSafe[coordinationv1.Lease](g, "lease", nsOption, "neco-cluster-exporter")
	pods := kubectlGetSafe[corev1.PodList](g, "pod", nsOption, "-l=app.kubernetes.io/name=neco-cluster-exporter")

	id := lease.Spec.HolderIdentity
	if id == nil {
		err := errors.New("holder identity is not set")
		g.Expect(err).NotTo(HaveOccurred())
	}

	for _, p := range pods.Items {
		if strings.HasPrefix(*id, p.Name) {
			return scrape(g, p.Status.PodIP+":8080")
		}
	}

	err := errors.New("no leader found")
	g.Expect(err).NotTo(HaveOccurred())
	return nil
}

func scrapeClusterNonLeader(g Gomega) []byte {
	lease := kubectlGetSafe[coordinationv1.Lease](g, "lease", nsOption, "neco-cluster-exporter")
	pods := kubectlGetSafe[corev1.PodList](g, "pod", nsOption, "-l=app.kubernetes.io/name=neco-cluster-exporter")

	id := lease.Spec.HolderIdentity
	if id == nil {
		err := errors.New("holder identity is not set")
		g.Expect(err).NotTo(HaveOccurred())
	}

	for _, p := range pods.Items {
		if !strings.HasPrefix(*id, p.Name) {
			return scrape(g, p.Status.PodIP+":8080")
		}
	}

	err := errors.New("no non-leader found")
	g.Expect(err).NotTo(HaveOccurred())
	return nil
}

func scrapeNode(g Gomega) []byte {
	return scrape(g, "neco-node-exporter.neco-exporter.svc")
}
