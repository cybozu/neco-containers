package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/cybozu/neco-containers/actions-slack-agent/slack"
	"github.com/spf13/cobra"
)

var isSucceded bool

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test is just for testing purpose",
	Long: `test is just for testing purpose.

This command sends a messaget to Slack Webhook.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("the number of arguments should be 1 but accepted %d", len(args))
		}

		webhookURL := args[0]
		c := slack.NewClient(webhookURL)
		return c.Notify(context.TODO(), "job", "namespace", "pod", isSucceded, time.Now())
	},
}

func init() {
	fs := testCmd.Flags()
	fs.BoolVarP(&isSucceded, "is-succeeded", "s", false, "send true message if success")
	rootCmd.AddCommand(testCmd)
}
