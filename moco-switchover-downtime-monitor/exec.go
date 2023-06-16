package main

import (
	"bytes"
	"context"
	"os/exec"
)

func execCmdWithInput(ctx context.Context, stdin []byte, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func execCmd(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	return execCmdWithInput(ctx, nil, name, args...)
}
