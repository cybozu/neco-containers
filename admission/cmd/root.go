package cmd

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/cybozu/neco-containers/admission/hooks"
	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

var config struct {
	metricsAddr           string
	probeAddr             string
	webhookAddr           string
	certDir               string
	httpProxyDefaultClass string
	configPath            string
	validImagePrefixes    []string
	imagePermissive       bool
	repositoryPermissive  bool
	zapOpts               zap.Options
}

var rootCmd = &cobra.Command{
	Use:   "admission",
	Short: "custom admission webhooks for Neco",
	Long:  `Custom admission webhooks for Neco.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		h, p, err := net.SplitHostPort(config.webhookAddr)
		if err != nil {
			return fmt.Errorf("invalid webhook address: %s, %v", config.webhookAddr, err)
		}
		numPort, err := strconv.Atoi(p)
		if err != nil {
			return fmt.Errorf("invalid webhook address: %s, %v", config.webhookAddr, err)
		}
		conf, err := parseConfig(config.configPath)
		if err != nil {
			return err
		}
		return run(h, numPort, conf)
	},
}

func parseConfig(configPath string) (*hooks.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var conf hooks.Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
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
	fs.StringVar(&config.probeAddr, "health-probe-addr", ":8081", "Listen address for health probes")
	fs.StringVar(&config.webhookAddr, "webhook-addr", ":9443", "Listen address for the webhook endpoint")
	fs.StringVar(&config.certDir, "cert-dir", "", "certificate directory")
	fs.StringVar(&config.httpProxyDefaultClass, "httpproxy-default-class", "", "Default Ingress class of HTTPProxy")
	fs.StringVar(&config.configPath, "config-path", "/etc/neco-admission/config.yaml", "Configuration for webhooks")
	fs.StringSliceVar(&config.validImagePrefixes, "valid-image-prefix", nil, "Valid prefixes of container images")
	config.imagePermissive = os.Getenv("VPOD_IMAGE_PERMISSIVE") == "true"
	config.repositoryPermissive = os.Getenv("VAPPLICATION_REPOSITORY_PERMISSIVE") == "true"

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	config.zapOpts.BindFlags(goflags)

	fs.AddGoFlagSet(goflags)
}
