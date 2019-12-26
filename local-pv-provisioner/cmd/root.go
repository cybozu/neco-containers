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
	pollingInterval  int
}

var rootCmd = &cobra.Command{
	Use:   "local-pv-provisioner",
	Short: "controller to create local PersistentVolume from device infos",
	Long:  `Controller to create local PersistentVolume from device infos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		config.metricsAddr = viper.GetString("metrics-addr")
		config.development = viper.GetBool("development")
		config.deviceDir = viper.GetString("device-dir")
		config.deviceNameFilter = viper.GetString("device-name-filter")
		config.nodeName = viper.GetString("node-name")
		config.pollingInterval = viper.GetInt("polling-interval")
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
	fs.String("metrics-addr", ":8080", "Listen address for metrics")
	fs.Bool("development", false, "Use development logger config")
	fs.String("device-dir", "/dev/disk/by-path/", "Path to the directory that stores the devices for which PersistentVolumes are created.")
	fs.String("device-name-filter", ".*", "A regular expression that allows selection of devices on device-idr to be created PersistentVolume.")
	fs.String("node-name", "", "The name of Node on which this program is running")
	fs.Uint("polling-interval", 10, "Polling interval to check devices. It is set by a second.")

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
