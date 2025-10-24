package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/components"
	"github.com/spf13/cobra"
)

var (
	bpfPerformanceInterval time.Duration
	// Add new components here

	log *slog.Logger
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
	cmd.Flags().DurationVar(&bpfPerformanceInterval, "bpf-performance-interval", time.Second*30, "Interval to check BPF performance")
	// Add new components here

	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func runMain() error {
	ctx := context.Background()

	go func() {
		if err := components.StartBPFPerformanceExporter(ctx, bpfPerformanceInterval); err != nil {
			panic(fmt.Errorf("failed to monitor BPF performance: %w", err))
		}
	}()
	// Add new components here

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	log.InfoContext(ctx, "start metrics server")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Error("metrics server stopped", slog.Any("error", err))
		return err
	}
	return nil
}
