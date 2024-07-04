package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/neco-containers/tcp-keepalive/internal/client"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run the tcp-keepalive client",
	Run:   runClient,
}

var clientCfg = &client.Config{}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.Flags().DurationVarP(&clientCfg.ReceiveTimeout, "timeout", "t", time.Second*15, "Deadline to receive a keepalive message")
	clientCmd.Flags().DurationVarP(&clientCfg.RetryInterval, "retry-interval", "r", time.Second, "Connect retry interval")
	clientCmd.Flags().IntVarP(&clientCfg.RetryNum, "retry", "y", 0, "Number of retries (-1 means infinite)")
	clientCmd.Flags().DurationVarP(&clientCfg.SendInterval, "interval", "i", time.Second*5, "Interval to send a keepalive message")
	clientCmd.Flags().StringVarP(&clientCfg.ServerAddr, "server", "s", "127.0.0.1:8000", "server address")
}

func runClient(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c, err := client.NewClient(clientCfg)
	if err != nil {
		log.Error("failed to create client", slog.Any("error", err))
		return
	}

	done := make(chan error)
	go func() {
		done <- c.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Info("Signal received. The client will be stopped after 5 seconds.")
		time.Sleep(5 * time.Second)
	case err := <-done:
		if err != nil {
			log.Error("client run failed", slog.Any("error", err))
		}
	}
}
