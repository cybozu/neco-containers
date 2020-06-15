package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/cobra"
)

var pushConfig struct {
	jobName  string
	addr     string
	interval time.Duration
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push metrics to Pushgateway",
	Long:  `Push metrics to Pushgateway`,
	Run: func(cmd *cobra.Command, args []string) {
		well.Go(pkg.NewWatcher(
			rootConfig.targetAddr,
			rootConfig.interval,
			&http.Client{},
		).Run)
		well.Go(func(ctx context.Context) error {
			pusher := push.New(pushConfig.addr, pushConfig.jobName).Gatherer(&prometheus.Registry{})
			tick := time.NewTicker(pushConfig.interval)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-tick.C:
					err := pusher.Add()
					if err != nil {
						log.Warn("push failed.", map[string]interface{}{
							"pushaddr": pushConfig.addr,
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
	fs.StringVarP(&pushConfig.addr, "push-addr", "", "", "Pushgateway addres.")
	fs.StringVarP(&pushConfig.jobName, "job-name", "", "", "Job name.")
	rootCmd.MarkPersistentFlagRequired("job-name")
	fs.DurationVarP(&pushConfig.interval, "push-interval", "", 10*time.Second, "Push interval.")
	rootCmd.MarkPersistentFlagRequired("job-name")

	rootCmd.AddCommand(pushCmd)
}
