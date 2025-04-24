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
	Long: `This command act to get TSR from iDRAC and put in the record onf Kintone:

Using the service tag of the server registered in the Kintone app as the key, 
find the IP address of the server's iDRAC/BMC (Baseboard Management Controller) from Sabakan, 
request a TSR (Technical Service Report) job, and register the obtained TSR in Kintone.`,

	Run: func(cmd *cobra.Command, args []string) {
		// iDRAC
		bc, err := bmc.LoadBMCUserConfig(cfgBmcUser)
		if err != nil {
			slog.Error("Can't read the config file of BMC", "err", err)
			os.Exit(1)
		}

		// Sabakan
		sa, _ := sabakan.ReadAppConfig(cfgSabakan)
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

		err = doMain(bc, sa, ka)
		if err != nil {
			fmt.Println("err=", err)
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
	rootCmd.PersistentFlags().StringVar(&cfgKintone, "kintone", "", "config file (default is $HOME/kintone.json)")
	rootCmd.PersistentFlags().StringVar(&cfgBmcUser, "bmc-user", "", "user file (default is /etc/bmc-user.json)")
	rootCmd.PersistentFlags().StringVar(&cfgSabakan, "sabakan", "", "config file (default is $HOME/sabakan.json)")
}
