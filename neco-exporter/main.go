package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/ciliumid"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/common/mock"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/server/bpf"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

const (
	scopeCommon  = "common"
	scopeCluster = "cluster"
	scopeServer  = "server"
)

type factory struct {
	scope        string
	metrixPrefix string
	newFunc      func() (exporter.Collector, error)
}

var (
	scope          string
	port           int
	collectorNames []string
	interval       time.Duration
)

var cmd = &cobra.Command{
	Use:   "neco-exporter",
	Short: "neco-exporter exposes node-local metrices",

	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMain()
	},
}

func init() {
	cmd.Flags().StringVar(&scope, "scope", scopeCluster, fmt.Sprintf("Collection scope (%s or %s)", scopeCluster, scopeServer))
	cmd.Flags().IntVar(&port, "port", 8080, "Specify port to expose metrics")
	cmd.Flags().StringSliceVar(&collectorNames, "collectors", []string{"bpf"}, "Specify collectors to activate")
	cmd.Flags().DurationVar(&interval, "interval", time.Second*30, "Interval to update metrics")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func runMain() error {
	factories := []factory{
		// scope: common
		{
			scope:        scopeCommon,
			metrixPrefix: "mock",
			newFunc:      mock.NewCollector,
		},
		// scope: cluster
		{
			scope:        scopeCluster,
			metrixPrefix: "ciliumid",
			newFunc:      ciliumid.NewCollector,
		},
		// scope: server
		{
			scope:        scopeServer,
			metrixPrefix: "bpf",
			newFunc:      bpf.NewCollector,
		},
	}

	collectors := make(map[string]exporter.Collector)
	for _, name := range collectorNames {
		index := slices.IndexFunc(factories, func(f factory) bool {
			return name == f.metrixPrefix
		})
		if index < 0 {
			return fmt.Errorf("unknown collector name: %s", name)
		}

		f := factories[index]
		switch {
		case scope == scopeCluster && f.scope == scopeServer:
			return fmt.Errorf("collector is not available in cluster-scope: %s", name)
		case scope == scopeServer && f.scope == scopeCluster:
			return fmt.Errorf("collector is not available in server-scope: %s", name)
		default:
			c, err := f.newFunc()
			if err != nil {
				return err
			}
			collectors[f.metrixPrefix] = c
		}
	}

	slog.Info("activate collectors", slog.Any("collectors", collectorNames))
	e := exporter.NewExporter(scope, port, collectors, interval)

	// controller-runtime will likely be needed in the near future,
	// so the dependency against it is not a problem
	ctx := ctrl.SetupSignalHandler()
	return e.Start(ctx)
}
