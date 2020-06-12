package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportConfig struct {
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Run server to export metrics for prometheus",
	Long:  `Run server to export metrics for prometheus`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("export: %#v", exportConfig)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
