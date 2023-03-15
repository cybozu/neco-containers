package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultConfigPath = "/etc/necosen/config.yaml"
)

var options struct {
	configFile  string
	addr        string
	reflection  bool
	logLevel    string
	tlsCertFile string
	tlsKeyFile  string
}

var rootCmd = &cobra.Command{
	Use:   "necosen",
	Short: "external authenticator",
	Long:  "external authenticator",
	RunE: func(cmd *cobra.Command, args []string) error {
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
	fs.StringVar(&options.configFile, "config-file", defaultConfigPath, "Configuration file path")
	fs.StringVar(&options.addr, "addr", ":50051", "gRPC endpoint")
	fs.BoolVar(&options.reflection, "reflection", false, "enable gRPC reflection")
	fs.StringVar(&options.logLevel, "log-level", "info", "zap log level")
	fs.StringVar(&options.tlsCertFile, "tls-cert-file", "", "cert path for TLS")
	fs.StringVar(&options.tlsKeyFile, "tls-key-file", "", "key path for TLS")
}
