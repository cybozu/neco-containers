package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/collector/bpf"
	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/collector/mock"
	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/exporter"
)

var (
	port           int
	collectorNames []string
	interval       time.Duration
	log            *slog.Logger
)

var cmd = &cobra.Command{
	Use:   "neco-server-exporter",
	Short: "neco-server-exporter exposes node-local metrices",

	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMain()
	},
}

func init() {
	cmd.Flags().IntVar(&port, "port", 8080, "Specify port to expose metrics")
	cmd.Flags().StringSliceVar(&collectorNames, "collectors", []string{"bpf"}, "Specify collectors to activate")
	cmd.Flags().DurationVar(&interval, "interval", time.Second*30, "Interval to update metrics")
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func runMain() error {
	candidates := []exporter.Collector{
		bpf.NewCollector(),
		mock.NewCollector(),
	}

	collectors := make([]exporter.Collector, 0)
	for _, name := range collectorNames {
		index := slices.IndexFunc(candidates, func(c exporter.Collector) bool {
			return name == c.MetricsPrefix()
		})
		if index < 0 {
			return fmt.Errorf("unknown collector name: %s", name)
		}
		collectors = append(collectors, candidates[index])
	}

	log.Info("activate collectors", slog.Any("collectors", collectorNames))
	e := exporter.NewExporter(port, collectors, interval)

	// controller-runtime will likely be needed in the near future,
	// so the dependency against it is not a problem
	ctx := ctrl.SetupSignalHandler()
	return e.Start(ctx)
}
