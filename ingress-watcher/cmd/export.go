package cmd

import (
	"context"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var exportConfig struct {
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Run server to export metrics for prometheus",
	Long:  `Run server to export metrics for prometheus`,
	Run: func(cmd *cobra.Command, args []string) {
		well.Go(func(ctx context.Context) error {
			return nil
		})

		well.Stop()
		err := well.Wait()
		if err != nil && !well.IsSignaled(err) {
			log.ErrorExit(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
