package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg/watch"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exportConfig struct {
	ListenAddr string
}

type logger struct{}

func (l logger) Println(v ...interface{}) {
	log.Error(fmt.Sprint(v...), nil)
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
		well.Go(watch.NewWatcher(
			rootConfig.TargetURLs,
			rootConfig.WatchInterval,
			&http.Client{},
		).Run)
		well.Go(func(ctx context.Context) error {
			mux := http.NewServeMux()
			handler := promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{
					ErrorLog:      logger{},
					ErrorHandling: promhttp.ContinueOnError,
				},
			)
			mux.Handle("/metrics", handler)
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
