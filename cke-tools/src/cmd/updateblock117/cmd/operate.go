package cmd

import (
	"github.com/cybozu-go/cke-tools/pkg/updateblock117"
	"github.com/spf13/cobra"
)

var operateCmd = &cobra.Command{
	Use:   "operate PV",
	Short: "move block device to new location and fix symbolic link",
	Long:  "move block device to new location and fix symbolic link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		pvName := args[0]
		existsOld, err := updateblock117.ExistsBlockDeviceAtOldLocation(pvName)
		if err != nil {
			return err
		}
		if existsOld {
			err = updateblock117.MoveBlockDeviceToTmp(pvName)
			if err != nil {
				return err
			}
		}

		existsTmp, err := updateblock117.ExistsBlockDeviceAtTmp(pvName)
		if err != nil {
			return err
		}
		if existsTmp {
			err = updateblock117.MoveBlockDeviceToNew(pvName)
			if err != nil {
				return err
			}
		}

		outdated, err := updateblock117.IsSymlinkOutdated(pvName)
		if err != nil {
			return err
		}
		if outdated {
			err = updateblock117.UpdateSymlink(pvName)
			if err != nil {
				return err
			}
		}

		return output("completed")
	},
}

func init() {
	rootCmd.AddCommand(operateCmd)
}
