package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/registry"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/option"
)

var cmd = &cobra.Command{
	Use:   "neco-exporter",
	Short: "neco-exporter exposes node-local metrics",

	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		return runMain()
	},
}

func init() {
	option.SetupOptionFlags(cmd)
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
	ctx := ctrl.SetupSignalHandler()

	candidates := registry.All()
	collectors := make([]exporter.Collector, 0)
	for _, name := range option.CollectorNames {
		index := slices.IndexFunc(candidates, func(c exporter.Collector) bool {
			return name == c.MetricsPrefix()
		})
		if index < 0 {
			return fmt.Errorf("unknown collector name: %s", name)
		}

		c := candidates[index]
		if option.Scope != c.Scope() {
			return fmt.Errorf("%s collector is not available in %s-scope", name, option.Scope)
		}
		if err := c.Setup(ctx); err != nil {
			return fmt.Errorf("failed to setup %s collector: %w", name, err)
		}

		collectors = append(collectors, c)
	}

	slog.Info("activate collectors", slog.Any("collectors", option.CollectorNames))
	e := exporter.NewExporter(option.Scope, option.Port, collectors, option.Interval)
	return e.Start(ctx)
}
