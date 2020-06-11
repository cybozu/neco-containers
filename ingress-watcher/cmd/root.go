package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var config struct {
	targetAddr string
}

var rootCmd = &cobra.Command{
	Use:   "ingress-watcher",
	Short: "Ingress monitoring tool for Neco",
	Long:  `Ingress monitoring tool for Neco.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%#v %#v", config, viper.GetString("target-addr"))
	},
}

// Execute executes the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&config.targetAddr, "target-addr", "", "", "Target Ingress address and port.")

	viper.BindPFlag("target-addr", fs.Lookup("target-addr"))
	fs.Set("target-addr", viper.GetString("target-addr"))

	rootCmd.MarkPersistentFlagRequired("target-addr")
}

func initConfig() {
	viper.SetEnvPrefix("iw")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
