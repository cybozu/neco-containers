package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg/common"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg/watch"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pushConfigFile string

var pushConfig struct {
	common.WatchConfig

	JobName      string
	PushAddr     string
	PushInterval time.Duration
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push metrics to Pushgateway",
	Long:  `Push metrics to Pushgateway`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pushConfigFile != "" {
			viper.SetConfigFile(pushConfigFile)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if err := viper.Unmarshal(&pushConfig); err != nil {
				return err
			}
		}

		if err := pushConfig.CheckCommonFlags(); err != nil {
			return err
		}

		if len(pushConfig.JobName) == 0 {
			return errors.New(`required flag "job-name" not set`)
		}

		if len(pushConfig.PushAddr) == 0 {
			return errors.New(`required flag "push-addr" not set`)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := pushConfig.GetClient()
		well.Go(watch.NewWatcher(
			pushConfig.TargetURLs,
			pushConfig.WatchInterval,
			&well.HTTPClient{Client: client},
		).Run)
		well.Go(func(ctx context.Context) error {
			tick := time.NewTicker(pushConfig.PushInterval)
			defer tick.Stop()

			pusher := push.New(pushConfig.PushAddr, pushConfig.JobName).Gatherer(registry)
			pusher.Client(client)
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-tick.C:
					err := pusher.Add()
					if err != nil {
						log.Warn("push failed.", map[string]interface{}{
							"addr":      pushConfig.PushAddr,
							log.FnError: err,
						})
					} else {
						log.Info("push succeeded.", map[string]interface{}{
							"addr": pushConfig.PushAddr,
						})
					}
				}
			}

		})
		well.Stop()
		err := well.Wait()
		if err != nil && !well.IsSignaled(err) {
			log.ErrorExit(err)
		}
	},
}

func init() {
	fs := pushCmd.Flags()
	pushConfig.SetCommonFlags(fs)
	fs.StringVarP(&pushConfigFile, "config", "", "", "Configuration YAML file path.")
	fs.StringVarP(&pushConfig.JobName, "job-name", "", "", "Job name.")
	fs.StringVarP(&pushConfig.PushAddr, "push-addr", "", "", "Pushgateway address.")
	fs.DurationVarP(&pushConfig.PushInterval, "push-interval", "", 10*time.Second, "Push interval.")

	rootCmd.AddCommand(pushCmd)
}
