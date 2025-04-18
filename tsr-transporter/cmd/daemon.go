/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var cfgIntervalSec int

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Daemon mode that acquire TSR from iDRAC and put it Kintone app",
	Long: `Daemon mode is the daemon process that get TSR from iDRAC and put in the record of Kintone.
	
Using the service tag of the server registered in the Kintone app as the key, 
find the IP address of the server's iDRAC/BMC (Baseboard Management Controller) from Sabakan, 
request a TSR (Technical Service Report) job, and register the obtained TSR in Kintone.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("daemon called")
		bc, sa, ka := readConfiguration(cmd, args)
		//
		err := jobLoopMain(bc, sa, ka)
		if err != nil {
			slog.Error("Error occurred in the job", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().IntVarP(&cfgIntervalSec, "interval", "i", 0, "Interval time (sec)")
}
