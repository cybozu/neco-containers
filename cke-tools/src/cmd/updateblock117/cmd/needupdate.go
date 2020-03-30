package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cybozu-go/cke-tools/pkg/updateblock117"
	"github.com/spf13/cobra"
)

var needUpdateCmd = &cobra.Command{
	Use:   "need-update PV",
	Short: "check that we should modify the path of the target device file or not",
	Long:  "check that we should modify the path of the target device file or not",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		pvName := args[0]

		existsOld, err := updateblock117.ExistsBlockDeviceAtOldLocation(pvName)
		if err != nil {
			return err
		}
		if existsOld {
			return output("yes")
		}

		existsTmp, err := updateblock117.ExistsBlockDeviceAtTmp(pvName)
		if err != nil {
			return err
		}
		if existsTmp {
			return output("yes")
		}

		outdated, err := updateblock117.IsSymlinkOutdated(pvName)
		if err != nil {
			return err
		}
		if outdated {
			return output("yes")
		}

		return output("no")
	},
}

func output(s string) error {
	res := struct {
		Result string `json:"result"`
	}{}
	res.Result = s
	out, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(out))
	return err
}

func init() {
	rootCmd.AddCommand(needUpdateCmd)
}
