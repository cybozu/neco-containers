package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/collector/bpf"
	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/exporter"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	interval time.Duration
	log      *slog.Logger
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
	e := exporter.NewExporter(interval)
	e.AddCollector(bpf.NewCollector())

	// controller-runtime will likely be needed in the near future,
	// so the dependency against it is not a problem
	ctx := ctrl.SetupSignalHandler()
	return e.Start(ctx)
}
