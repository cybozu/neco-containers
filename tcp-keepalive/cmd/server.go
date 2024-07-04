package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/neco-containers/tcp-keepalive/internal/server"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the tcp-keepalive server",
	Run:   runServer,
}

var serverCfg = &server.Config{}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverCfg.ListenAddr, "listen", "l", "127.0.0.1:8000", "Listen address and port")
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	s, err := server.NewServer(serverCfg)
	if err != nil {
		log.Error("failed to create server", slog.Any("error", err))
		return
	}

	done := make(chan error)
	go func() {
		done <- s.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Info("Signal received. The server will be stopped after 5 seconds.")
		time.Sleep(5 * time.Second)
	case err := <-done:
		if err != nil {
			log.Error("server run failed", slog.Any("error", err))
		}
	}
}
