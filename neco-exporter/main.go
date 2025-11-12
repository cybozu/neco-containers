package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/registry"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

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
	cmd.Flags().StringVar(&scope, "scope", constants.ScopeCluster, fmt.Sprintf("Collection scope (%s or %s)", constants.ScopeCluster, constants.ScopeNode))
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
	candidates := registry.All()
	collectors := make([]exporter.Collector, 0)
	for _, name := range collectorNames {
		index := slices.IndexFunc(candidates, func(c exporter.Collector) bool {
			return name == c.MetricsPrefix()
		})
		if index < 0 {
			return fmt.Errorf("unknown collector name: %s", name)
		}

		c := candidates[index]
		if scope != c.Scope() {
			return fmt.Errorf("%s collector is not available in %s-scope", name, scope)
		}
		if err := c.Setup(); err != nil {
			return fmt.Errorf("failed to setup %s collector: %w", name, err)
		}

		collectors = append(collectors, c)
	}

	slog.Info("activate collectors", slog.Any("collectors", collectorNames))
	e := exporter.NewExporter(scope, port, collectors, interval)

	// controller-runtime will likely be needed in the near future,
	// so the dependency against it is not a problem
	ctx := ctrl.SetupSignalHandler()
	return e.Start(ctx)
}
