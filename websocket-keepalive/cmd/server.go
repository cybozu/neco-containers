package cmd

import (
	"log/slog"
	"os"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
	"github.com/cybozu/neco-containers/websocket-keepalive/internal/server"

	"github.com/spf13/cobra"
)

var serverConfig = server.Config{}
var serverMetricsConfig = &metrics.Config{}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start WebSocket server",
	Long:  "Start a WebSocket server that handles client connections",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting WebSocket server", "host", serverConfig.Host, "port", serverConfig.Port)
		if err := server.RunWithConfig(&serverConfig, serverMetricsConfig); err != nil {
			slog.Error("Failed to run WebSocket server", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&serverConfig.Host, "listen", "l", "0.0.0.0", "Host to listen on")
	serverCmd.Flags().IntVarP(&serverConfig.Port, "port", "p", 9000, "Port to listen on")
	serverCmd.Flags().BoolVarP(&serverMetricsConfig.Export, "metrics", "m", true, "Enable metrics")
	serverCmd.Flags().StringVarP(&serverMetricsConfig.AddrPort, "metrics-server", "a", "0.0.0.0:8081", "Metrics server address and port")
}
