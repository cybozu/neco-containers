package cmd

import (
	"fmt"
	"net"
	"net/netip"
	"os"

	"github.com/cybozu/neco-containers/necosen/pkg/config"
	"github.com/cybozu/neco-containers/necosen/pkg/server"
	"go.uber.org/zap"
)

func run() error {
	zapConfig := zap.NewProductionConfig()
	level, err := zap.ParseAtomicLevel(options.logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	zapConfig.Level = level
	logger, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	// Load Config
	cfgData, err := os.ReadFile(options.configFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", options.configFile, err)
	}
	cfg := &config.Config{}
	if err := cfg.Load(cfgData); err != nil {
		return fmt.Errorf("unable to load the configuration file: %w", err)
	}

	var allowedCIDRs []netip.Prefix
	if cfg.SourceIP.AllowedCIDRs != nil {
		for _, cidr := range cfg.SourceIP.AllowedCIDRs {
			allowedCIDRs = append(allowedCIDRs, netip.MustParsePrefix(cidr))
		}
	}

	l, err := net.Listen("tcp", options.addr)
	if err != nil {
		return err
	}

	a := server.NewAuthorizationServer(server.Config{
		Listener:     l,
		Reflection:   options.reflection,
		TLSCertFile:  options.tlsCertFile,
		TLSKeyFile:   options.tlsKeyFile,
		AllowedCIDRs: allowedCIDRs,
	}, logger)
	return a.Start()
}
