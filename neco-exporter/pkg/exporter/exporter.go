package exporter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type Exporter struct {
	scope          string
	port           int
	metrixPrefixes []string
	collectors     []Collector
	interval       time.Duration
}

func NewExporter(scope string, port int, collectors map[string]Collector, interval time.Duration) *Exporter {
	metrixPrefixes := make([]string, 0, len(collectors))
	collectorList := make([]Collector, 0, len(collectors))
	for k, v := range collectors {
		metrixPrefixes = append(metrixPrefixes, k)
		collectorList = append(collectorList, v)
	}

	return &Exporter{
		scope:          scope,
		port:           port,
		metrixPrefixes: metrixPrefixes,
		collectors:     collectorList,
		interval:       interval,
	}
}

func (e *Exporter) Start(ctx context.Context) error {
	go func() {
		if err := e.Run(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to scrape metrics", slog.Any("error", err))
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%d", e.port)
	slog.InfoContext(ctx, "start metrics server", slog.Any("address", addr))

	server := http.Server{
		Addr: addr,
	}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	serveErr := make(chan error)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "metrics server stopped", slog.Any("error", err))
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		slog.InfoContext(ctx, "shutting down server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// Run starts the metric collection loop.
// The error return is reserved for future fatal error conditions.
// Currently always returns nil except when context is cancelled.
func (e *Exporter) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	prev := make([]map[string]struct{}, len(e.collectors))
	health := make([]bool, len(e.collectors))
	dur := make([]time.Duration, len(e.collectors))

	for {
		var wg sync.WaitGroup
		wg.Add(len(e.collectors))
		for i, c := range e.collectors {
			go func(i int) {
				next := make(map[string]struct{})

				startTime := time.Now()
				r, err := c.Collect(ctx)
				health[i] = (err == nil)
				dur[i] = time.Since(startTime)

				prefix := e.metrixPrefixes[i]
				if err != nil {
					slog.ErrorContext(ctx, "failed to collect metrics", slog.Any("error", err), slog.String("collector", e.metrixPrefixes[i]))
				} else {
					for _, m := range r {
						n := BuildMetricName(e.scope, prefix, m.Name, m.Labels)
						counter := metrics.GetOrCreateFloatCounter(n)
						counter.Set(m.Value)

						delete(prev[i], n)
						next[n] = struct{}{}
					}
				}

				// Remove stale metrics
				for k := range prev[i] {
					metrics.UnregisterMetric(k)
				}
				prev[i] = next

				wg.Done()
			}(i)
		}
		wg.Wait()

		const collectorPrefix = "collector"
		for i := 0; i < len(e.collectors); i++ {
			labels := map[string]string{
				"collector": e.metrixPrefixes[i],
			}

			n := BuildMetricName(e.scope, collectorPrefix, "health", labels)
			counter := metrics.GetOrCreateFloatCounter(n)
			if health[i] {
				counter.Set(1)
			} else {
				counter.Set(0)
			}

			n = BuildMetricName(e.scope, collectorPrefix, "process_seconds", labels)
			counter = metrics.GetOrCreateFloatCounter(n)
			counter.Set(dur[i].Seconds())
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
