package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var rootConfig struct {
	targetAddr string
	interval   time.Duration
}

var rootCmd = &cobra.Command{
	Use:   "ingress-watcher",
	Short: "Ingress monitoring tool for Neco",
	Long:  `Ingress monitoring tool for Neco.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%#v", rootConfig)
	},
}

// Execute executes the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&rootConfig.targetAddr, "target-addr", "", "", "Target Ingress address and port.")
	rootCmd.MarkPersistentFlagRequired("target-addr")
	fs.DurationVarP(&rootConfig.interval, "interval", "", 5*time.Second, "Polling interval.")

	prometheus.MustRegister(
		metrics.HTTPGetSuccessfulTotal,
		metrics.HTTPGetFailTotal,
		metrics.HTTPSGetSuccessfulTotal,
		metrics.HTTPSGetFailTotal,
	)
}
