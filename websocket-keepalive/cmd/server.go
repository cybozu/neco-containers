package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/server"

	"github.com/spf13/cobra"
)

var serverConfig = server.Config{}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start WebSocket server",
	Long:  "Start a WebSocket server that handles client connections",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting WebSocket server", "host", serverConfig.Host, "port", serverConfig.Port)
		if err := server.Run(serverConfig.Host, serverConfig.Port, serverConfig.PingInterval); err != nil {
			slog.Error("Failed to run WebSocket server", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&serverConfig.Host, "listen", "l", "0.0.0.0", "Host to listen on")
	serverCmd.Flags().IntVarP(&serverConfig.Port, "port", "p", 9000, "Port to listen on")
	serverCmd.Flags().DurationVarP(&serverConfig.PingInterval, "ping-interval", "i", 5*time.Second, "Ping interval in seconds")
}
