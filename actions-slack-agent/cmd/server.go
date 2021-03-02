package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/actions-slack-agent/agent"
	"github.com/cybozu/neco-containers/actions-slack-agent/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverConfig struct {
	listenAddr string
	listenPort int
	webhookURL string
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "server starts Slack agent",
	Long: `server starts Slack agent

In addition to flags, the following environment variables are read:
	ACTIONS_LISTEN_ADDR          Listening address
	ACTIONS_LISTEN_PORT          Listening port number
	ACTIONS_WEBHOOK_URL          Slack Webhook URL
`,
	Run: func(cmd *cobra.Command, args []string) {
		url := viper.GetString("webhook-url")
		if len(url) == 0 {
			log.ErrorExit(errors.New(`"webhook-url" flag should not be empty`))
		}

		env := well.NewEnvironment(context.Background())
		s := agent.NewServer(
			serverConfig.listenAddr,
			serverConfig.listenPort,
			slack.NewClient(serverConfig.webhookURL),
		)
		env.Go(s.Start)
		err := well.Wait()
		if err != nil && !well.IsSignaled(err) {
			log.ErrorExit(err)
		}
	},
}

func init() {
	fs := serverCmd.Flags()
	fs.StringVarP(&serverConfig.listenAddr, "listen-addr", "a", "0.0.0.0", "Listening address")
	fs.IntVarP(&serverConfig.listenPort, "listen-port", "p", 8080, "Listening port number")
	fs.StringVarP(&serverConfig.webhookURL, "webhook-url", "u", "", "Slack Webhook URL")
	rootCmd.AddCommand(serverCmd)

	viper.SetEnvPrefix("actions")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
