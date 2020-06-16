package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var registry *prometheus.Registry
var configFile string

var rootConfig struct {
	TargetURLs    []string
	WatchInterval time.Duration
}

var rootCmd = &cobra.Command{
	Use:   "ingress-watcher",
	Short: "Ingress monitoring tool for Neco",
	Long:  `Ingress monitoring tool for Neco.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if err := viper.Unmarshal(&rootConfig); err != nil {
				return err
			}
		}

		if len(rootConfig.TargetURLs) == 0 {
			return errors.New("required flag \"target-urls\" not set")
		}

		return nil
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
	fs.StringArrayVarP(&rootConfig.TargetURLs, "target-urls", "", nil, "Target Ingress address and port.")
	fs.DurationVarP(&rootConfig.WatchInterval, "watch-interval", "", 5*time.Second, "Watching interval.")
	fs.StringVarP(&configFile, "config", "", "", "Configuration YAML file path.")

	registry = prometheus.NewRegistry()
	registry.MustRegister(
		metrics.HTTPGetSuccessfulTotal,
		metrics.HTTPGetFailTotal,
	)
}
