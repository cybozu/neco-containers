package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exportConfig struct {
	ListenAddr string
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Run server to export metrics for prometheus",
	Long:  `Run server to export metrics for prometheus`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if configFile != "" {
			if err := viper.Unmarshal(&exportConfig); err != nil {
				return err
			}
		}

		if len(exportConfig.ListenAddr) == 0 {
			return errors.New("required flag \"listen-addr\" not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		well.Go(pkg.NewWatcher(
			rootConfig.TargetAddrs,
			rootConfig.Interval,
			&http.Client{},
		).Run)
		well.Go(func(ctx context.Context) error {
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			serv := &well.HTTPServer{
				Server: &http.Server{
					Addr:    exportConfig.ListenAddr,
					Handler: mux,
				},
			}
			return serv.ListenAndServe()
		})
		well.Stop()
		err := well.Wait()
		if err != nil && !well.IsSignaled(err) {
			log.ErrorExit(err)
		}
	},
}

func init() {
	fs := exportCmd.Flags()
	fs.StringVarP(&exportConfig.ListenAddr, "listen-addr", "", "0.0.0.0:8080", "Listen address of metrics server.")

	rootCmd.AddCommand(exportCmd)
}
