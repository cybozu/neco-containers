package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/pkg/watch"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exportConfigFile string

var exportConfig struct {
	TargetURLs     []string
	WatchInterval  time.Duration
	ListenAddr     string
	PermitInsecure bool
	ResolveRules   []string
	// TODO
}

type logger struct{}

func (l logger) Println(v ...interface{}) {
	log.Error(fmt.Sprint(v...), nil)
}

// `ingres-watcher export` is not used in neco-apps, but we leave it here.
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Run server to export metrics for prometheus",
	Long:  `Run server to export metrics for prometheus`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		exportConfig.ResolveRules = make(map[string]interface{})
		for _, rule := range resolveRules {
			split := strings.Split(rule, ":")
			if len(split) != 2 {
				return errors.New(`invalid format in "resolve-rules" : ` + rule)
			}
			exportConfig.ResolveRules[split[0]] = split[1]
		}

		if exportConfigFile != "" {
			viper.SetConfigFile(exportConfigFile)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if err := viper.Unmarshal(&exportConfig); err != nil {
				return err
			}
		}

		if len(exportConfig.TargetURLs) == 0 {
			return errors.New(`required flag "target-urls" not set`)
		}

		if len(exportConfig.ListenAddr) == 0 {
			return errors.New(`required flag "listen-addr" not set`)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := &http.Client{}
		if exportConfig.PermitInsecure {
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		well.Go(watch.NewWatcher(
			exportConfig.TargetURLs,
			exportConfig.WatchInterval,
			&well.HTTPClient{Client: client},
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
	fs.StringArrayVarP(&exportConfig.TargetURLs, "target-urls", "", nil, "Target Ingress address and port.")
	fs.DurationVarP(&exportConfig.WatchInterval, "watch-interval", "", 5*time.Second, "Watching interval.")
	fs.StringVarP(&exportConfigFile, "config", "", "", "Configuration YAML file path.")
	fs.BoolVar(&exportConfig.PermitInsecure, "permit-insecure", false, "Permit insecure access to targets.")
	fs.StringArrayVarP(&exportConfig.ResolveRules, "resolve-rules", "", nil, "Resolve rules from fqdn to IPv4 address (ex. example.com:192.168.0.1).")

	rootCmd.AddCommand(exportCmd)
}
