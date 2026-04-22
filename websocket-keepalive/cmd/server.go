package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := server.RunWithConfig(ctx, &serverConfig, serverMetricsConfig); err != nil {
			slog.Error("Failed to run WebSocket server", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&serverConfig.Host, "listen", "l", "0.0.0.0", "Host to listen on")
	serverCmd.Flags().IntVarP(&serverConfig.Port, "port", "p", 9000, "Port to listen on")
	serverCmd.Flags().DurationVarP(&serverConfig.PingInterval, "ping-interval", "i", 10*time.Second, "Expected client ping interval (used to compute read deadline)")
	serverCmd.Flags().BoolVarP(&serverMetricsConfig.Export, "metrics", "m", true, "Enable metrics")
	serverCmd.Flags().StringVarP(&serverMetricsConfig.AddrPort, "metrics-server", "a", "0.0.0.0:8081", "Metrics server address and port")
}
