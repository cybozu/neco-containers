package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var config struct {
	metricsAddr            string
	nodeName               string
	development            bool
	pollingInterval        time.Duration
	zapOpts                zap.Options
	defaultPVSpecConfigMap string
	namespaceName          string
}

var rootCmd = &cobra.Command{
	Use:   "local-pv-provisioner",
	Short: "controller to create local PersistentVolume from device infos",
	Long:  `Controller to create local PersistentVolume from device infos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		config.metricsAddr = viper.GetString("metrics-addr")
		config.development = viper.GetBool("development")
		config.defaultPVSpecConfigMap = viper.GetString("default-pv-spec-configmap")
		config.nodeName = viper.GetString("node-name")
		config.pollingInterval = viper.GetDuration("polling-interval")
		config.namespaceName = viper.GetString("namespace-name")
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
	fs.String("node-name", "", "The name of Node on which this program is running")
	fs.Duration("polling-interval", 5*time.Minute, "Polling interval to check devices.")
	fs.String("default-pv-spec-configmap", "", "A ConfigMap name that should be used if the Node doesn't have the pv-spec-configmap annotation.")
	fs.String("namespace-name", "", "the name of the namespace in which this program is running")

	if err := viper.BindPFlags(fs); err != nil {
		panic(err)
	}
	viper.SetEnvPrefix("lp")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	config.zapOpts.BindFlags(goflags)

	fs.AddGoFlagSet(goflags)
}
