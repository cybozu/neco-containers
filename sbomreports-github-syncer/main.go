package main

import (
	"fmt"
	"os"

	"github.com/cybozu/neco-containers/sbomreports-github-syncer/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
