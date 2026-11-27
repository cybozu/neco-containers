package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"

	"github.com/spf13/pflag"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	"github.com/cybozu/neco-containers/admission/hooks"
)

var config struct {
	metricsAddr                string
	probeAddr                  string
	webhookAddr                string
	certDir                    string
	httpProxyDefaultClass      string
	configPath                 string
	validImagePrefixes         []string
	imagePermissive            bool
	ephemeralStoragePermissive bool
	cpuRequestReducerEnabled   bool
	zapOpts                    zap.Options
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

func main() {
	fs := pflag.NewFlagSet("neco-admission", pflag.ExitOnError)
	fs.StringVar(&config.metricsAddr, "metrics-addr", ":8080", "Listen address for metrics")
	fs.StringVar(&config.probeAddr, "health-probe-addr", ":8081", "Listen address for health probes")
	fs.StringVar(&config.webhookAddr, "webhook-addr", ":9443", "Listen address for the webhook endpoint")
	fs.StringVar(&config.certDir, "cert-dir", "", "certificate directory")
	fs.StringVar(&config.httpProxyDefaultClass, "httpproxy-default-class", "", "Default Ingress class of HTTPProxy")
	fs.StringVar(&config.configPath, "config-path", "/etc/neco-admission/config.yaml", "Configuration for webhooks")
	fs.StringSliceVar(&config.validImagePrefixes, "valid-image-prefix", nil, "Valid prefixes of container images")

	goflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(goflags)
	config.zapOpts.BindFlags(goflags)
	fs.AddGoFlagSet(goflags)

	// The error is ignored because fs is created with pflag.ExitOnError.
	// Parse never returns an error; it exits the process instead.
	_ = fs.Parse(os.Args[1:])

	// Drop empty prefixes such as the one in "a/,", because an empty prefix matches every image.
	config.validImagePrefixes = slices.DeleteFunc(config.validImagePrefixes, func(p string) bool { return p == "" })

	config.imagePermissive = os.Getenv("VPOD_IMAGE_PERMISSIVE") == "true"
	config.ephemeralStoragePermissive = os.Getenv("VPOD_EPHEMERAL_STORAGE_PERMISSIVE") == "true"
	config.cpuRequestReducerEnabled = os.Getenv("VPOD_CPU_REQUEST_REDUCE_ENABLE") == "true"

	if err := subMain(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func subMain() error {
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
}
