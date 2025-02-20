package cmd

import (
	"os"

	"github.com/cybozu/neco-containers/lbip-manager/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.GetLogger()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lbip-manager",
	Short: "lbip-manager manages IP addresses for Kubernetes LoadBalancers",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
