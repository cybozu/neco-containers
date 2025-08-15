package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/client"
	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"

	"github.com/spf13/cobra"
)

var clientConfig = client.Config{}
var clientMetricsConfig = &metrics.Config{}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start WebSocket client",
	Long:  "Start a WebSocket client that sends periodic ping messages",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting WebSocket client", "host", clientConfig.Host, "port", clientConfig.Port)
		if err := client.RunWithConfig(&clientConfig, clientMetricsConfig); err != nil {
			slog.Error("Failed to run WebSocket client", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	clientCmd.Flags().StringVarP(&clientConfig.Host, "host", "H", "localhost", "Server host to connect to")
	clientCmd.Flags().IntVarP(&clientConfig.Port, "port", "p", 9000, "Server port to connect to")
	clientCmd.Flags().DurationVarP(&clientConfig.PingInterval, "ping-interval", "i", 10 * time.Second, "Interval for sending ping messages")
	clientCmd.Flags().IntVarP(&clientConfig.MaxPingRetries, "max-retry-limit", "r", 3, "Limit for retrying to send ping")
	clientCmd.Flags().BoolVarP(&clientMetricsConfig.Export, "metrics", "m", true, "Enable metrics")
	clientCmd.Flags().StringVarP(&clientMetricsConfig.AddrPort, "metrics-server", "a", "0.0.0.0:8080", "Metrics server address and port")

}
