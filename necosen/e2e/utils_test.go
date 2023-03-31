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
	yqPath      = "../bin/yq"
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
	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), stderr.Bytes(), nil
	}
	_, file := filepath.Split(path)
	return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("%s failed with %s: stderr=%s", file, err, stderr)
}

func kubectl(input []byte, args ...string) ([]byte, []byte, error) {
	args = append([]string{"--context", "kind-necosen"}, args...)
	return runCommand(kubectlPath, input, args...)
}

func kubectlSafe(input []byte, args ...string) []byte {
	out, _, err := kubectl(input, args...)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return out
}

func yq(input []byte, args ...string) ([]byte, []byte, error) {
	return runCommand(yqPath, input, args...)
}

func yqSafe(input []byte, args ...string) []byte {
	out, _, err := yq(input, args...)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return out
}
