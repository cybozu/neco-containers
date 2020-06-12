package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootConfig struct {
	targetAddr string
}

var rootCmd = &cobra.Command{
	Use:   "ingress-watcher",
	Short: "Ingress monitoring tool for Neco",
	Long:  `Ingress monitoring tool for Neco.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%#v", rootConfig)
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
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&rootConfig.targetAddr, "target-addr", "", "", "Target Ingress address and port.")
	rootCmd.MarkPersistentFlagRequired("target-addr")
}
