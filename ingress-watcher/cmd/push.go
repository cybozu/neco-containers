package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushConfig struct {
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push metrics to Pushgateway",
	Long:  `Push metrics to Pushgateway`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("push: %#v", pushConfig)
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
