package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var log *slog.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tcp-keepalive",
	Short: "tcp-keepalive is a simple TCP server and client program to confirm the long live connectivity.",
}

func init() {
	initLogger()
}

func initLogger() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
