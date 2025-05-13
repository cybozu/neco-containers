package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of this command",
	Long:  `show version of this command`,
	Run: func(cmd *cobra.Command, args []string) {
		fp, err := os.Open("TAG")
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		buf := make([]byte, 1024)
		n, err := fp.Read(buf)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			panic(err)
		}
		fmt.Print("Version:", string(buf))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
