package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"
)

var config struct {
	metricsAddr      string
	nodeName         string
	deviceDir        string
	deviceNameFilter string
	development      bool
}

var rootCmd = &cobra.Command{
	Use:   "local-pv-provisioner",
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
	fs.StringVar(&config.deviceDir, "device-dir", "/dev/disk/by-path/", "Path to the directory that stores the devices for which PersistentVolumes are created.")
	fs.StringVar(&config.deviceNameFilter, "device-name-filter", ".*", "A regular expression that allows selection of devices on device-idr to be created PersistentVolume.")
	fs.StringVar(&config.nodeName, "node-name", "", "The name of Node on which this program is running")

	if err := viper.BindPFlags(fs); err != nil {
		panic(err)
	}
	viper.SetEnvPrefix("lp")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	fs.AddGoFlagSet(goflags)
}
