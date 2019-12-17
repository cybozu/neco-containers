package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var config struct {
	metricsAddr string
	development bool
}

var rootCmd = &cobra.Command{
	Use:   "local-pv-creator",
	Short: "controller to create local PersistentVolume from device infos",
	Long:  `Controller to create local PersistentVolume from device infos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return run()
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
	fs := rootCmd.Flags()
	fs.StringVar(&config.metricsAddr, "metrics-addr", ":8080", "Listen address for metrics")
	fs.BoolVar(&config.development, "development", false, "Use development logger config")

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	fs.AddGoFlagSet(goflags)
}
