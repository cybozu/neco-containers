package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
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
		if exportConfigFile != "" {
			viper.SetConfigFile(exportConfigFile)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			if err := viper.Unmarshal(&exportConfig); err != nil {
				return err
			}
		}

		for _, rule := range exportConfig.ResolveRules {
			split := strings.Split(rule, ":")
			if len(split) != 2 {
				return errors.New(`invalid format in "resolve-rules" : ` + rule)
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
		var transport *http.Transport
		if exportConfig.PermitInsecure {
			if transport == nil {
				transport = &http.Transport{}
			}
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		if len(exportConfig.ResolveRules) > 0 {
			resolveMap := make(map[string]string)
			for _, rules := range exportConfig.ResolveRules {
				s := strings.Split(rules, ":")
				resolveMap[s[0]] = s[1]
			}

			dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				splitAddr := strings.Split(address, ":")
				if len(splitAddr) > 2 {
					return nil, errors.New(`invalid format : ` + address)
				}

				if ip, ok := resolveMap[splitAddr[0]]; ok {
					return d.DialContext(ctx, network, ip+":"+splitAddr[1])
				}
				return d.DialContext(ctx, network, address)
			}

			if transport == nil {
				transport = &http.Transport{}
			}
			transport.DialContext = dialerFunc
		}

		client := &http.Client{}
		if transport != nil {
			client.Transport = transport
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
	fs.StringArrayVarP(&exportConfig.ResolveRules, "resolve-rules", "", nil, "Resolve rules from FQDN to IPv4 address (ex. example.com:192.168.0.1).")

	rootCmd.AddCommand(exportCmd)
}
