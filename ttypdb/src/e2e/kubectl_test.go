package e2e

import (
	"bytes"
	"io"
	"os/exec"
)

func kubectl(args ...string) ([]byte, []byte, error) {
	return kubectlWithReaderStdin(nil, args...)
}

func kubectlWithReaderStdin(stdin io.Reader, args ...string) ([]byte, []byte, error) {
	stdoutBuffer := new(bytes.Buffer)
	stderrBuffer := new(bytes.Buffer)
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err := cmd.Run()
	return stdoutBuffer.Bytes(), stderrBuffer.Bytes(), err
}
