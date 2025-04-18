package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var cfgKintone string
var cfgSabakan string
var cfgBmcUser string

var rootCmd = &cobra.Command{
	Use:   "tsr-transporter",
	Short: "Acquire TSR from iDRAC and put it Kintone app",
	Long: `This command act to get TSR from iDRAC and put in the record of Kintone:

Using the service tag of the server registered in the Kintone app as the key, 
find the IP address of the server's iDRAC/BMC (Baseboard Management Controller) from Sabakan, 
request a TSR (Technical Service Report) job, and register the obtained TSR in Kintone.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(cfgKintone) == 0 && len(cfgSabakan) == 0 && len(cfgBmcUser) == 0 {
			fmt.Println("Error: Must set flag or sub-command")
			fmt.Println()
			cmd.Help()
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancelCause := context.WithCancelCause(context.Background())
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-c
			slog.Info("Catch SIGNAL", "signal number=", sig)
			cancelCause(fmt.Errorf("%v", sig))
			os.Exit(1)
		}()

		bc, sa, ka := readConfiguration(cmd, args)
		err := jobMain(ctx, bc, sa, ka)
		if err != nil {
			slog.Error("Error occurred in the job", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgKintone, "kintone", "k", "", "Kintone App config")
	rootCmd.PersistentFlags().StringVarP(&cfgBmcUser, "bmc-user", "b", "", "BMC user config")
	rootCmd.PersistentFlags().StringVarP(&cfgSabakan, "sabakan", "s", "", "Sabakan config")
}
