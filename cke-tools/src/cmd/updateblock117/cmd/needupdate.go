package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cybozu-go/cke-tools/pkg/updateblock117"
	"github.com/spf13/cobra"
)

var needUpdateCmd = &cobra.Command{
	Use:   "need-update <block-pv-name>",
	Short: "check that we should modify the path of the target device file or not",
	Long: "check that we should modify the path of the target device file or not",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pvName := args[0]
		res := result{}
		existsOld, err := updateblock117.ExistsBlockDeviceAtOldLocation(pvName)
		if err != nil {
			return err
		}
		if existsOld {
			res.Result = "yes"
			_, err = fmt.Println(json.Marshal(res))
			return err
		}

		existsTmp, err := updateblock117.ExistsBlockDeviceAtTmp(pvName)
		if err != nil {
			return err
		}
		if existsTmp {
			res.Result = "yes"
			_, err = fmt.Println(json.Marshal(res))
			return err
		}

		outdated, err := updateblock117.IsSymlinkOutdated(pvName)
		if err != nil {
			return err
		}
		if outdated {
			res.Result = "yes"
			_, err = fmt.Println(json.Marshal(res))
			return err
		}
		return nil
	},
}

type result struct {
	Result string	`json:"result"`
}
