package cmd

import (
	"context"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var exportConfig struct {
	listenAddr string
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Run server to export metrics for prometheus",
	Long:  `Run server to export metrics for prometheus`,
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
					Addr:    exportConfig.listenAddr,
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
	fs.StringVarP(&exportConfig.listenAddr, "listen-addr", "", "0.0.0.0:8080", "Listen address of metrics server.")

	rootCmd.AddCommand(exportCmd)
}
