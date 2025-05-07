package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cybozu/neco-containers/tsr-transporter/bmc"
	kintone "github.com/cybozu/neco-containers/tsr-transporter/kintone"
	"github.com/cybozu/neco-containers/tsr-transporter/sabakan"
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
		// iDRAC
		bc, err := bmc.ReadUsers(cfgBmcUser)
		if err != nil {
			slog.Error("Can't read the config file of BMC", "err", err)
			os.Exit(1)
		}

		// Sabakan
		sa, _ := sabakan.ReadConfig(cfgSabakan)
		if err != nil {
			slog.Error("Can't read the config file of sabakan", "err", err)
			os.Exit(1)
		}

		// Kintone
		ka, err := kintone.ReadAppConfig(cfgKintone)
		if err != nil {
			slog.Error("Can't read the config file of kintone", "err", err)
			os.Exit(1)
		}

		err = jobMain(bc, sa, ka)
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
