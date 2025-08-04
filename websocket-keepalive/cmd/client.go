package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/client"

	"github.com/spf13/cobra"
)

var clientConfig = client.Config{}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start WebSocket client",
	Long:  "Start a WebSocket client that sends periodic ping messages",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting WebSocket client", "host", clientConfig.Host, "port", clientConfig.Port)
		if err := client.Run(clientConfig.Host, clientConfig.Port, clientConfig.PingInterval); err != nil {
			slog.Error("Failed to run WebSocket client", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	clientCmd.Flags().StringVarP(&clientConfig.Host, "host", "H", "localhost", "Server host to connect to")
	clientCmd.Flags().IntVarP(&clientConfig.Port, "port", "p", 9000, "Server port to connect to")
	clientCmd.Flags().DurationVarP(&clientConfig.PingInterval, "ping-interval", "i", 10 * time.Second, "Interval for sending ping messages")
}
