/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// mutatingCmd represents the mutating command
var mutatingCmd = &cobra.Command{
	Use:   "mutating",
	Short: "Run mutating webhook server which modifies the spec.loadBalancerIP field",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mutating called")
	},
}

func init() {
	rootCmd.AddCommand(mutatingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mutatingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mutatingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
