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
	scope      string
	port       int
	collectors []Collector
	interval   time.Duration
}

func NewExporter(scope string, port int, collectors []Collector, interval time.Duration) *Exporter {
	return &Exporter{
		scope:      scope,
		port:       port,
		collectors: collectors,
		interval:   interval,
	}
}

func (e *Exporter) Start(ctx context.Context) (ret error) {
	ctx, cancelCause := context.WithCancelCause(ctx)
	defer cancelCause(nil)

	go func() {
		if err := e.Run(ctx); err != nil {
			err = fmt.Errorf("failed to scrape metrics: %w", err)
			cancelCause(err)
		} else {
			cancelCause(nil)
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%d", e.port)
	slog.InfoContext(ctx, "start metrics server", slog.Any("address", addr))

	server := http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("metrics server stopped: %w", err)
			cancelCause(err)
		} else {
			cancelCause(nil)
		}
	}()
	defer func() {
		slog.InfoContext(ctx, "shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := server.Shutdown(shutdownCtx)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			err = nil
		}

		if err != nil {
			if ret != nil {
				// Log previously-captured error before overwrite
				slog.ErrorContext(ctx, "server failed to shutdown after exporter failure", slog.Any("error", ret))
			}
			ret = err
		}
	}()

	<-ctx.Done()
	ret = context.Cause(ctx)
	if errors.Is(ret, context.Canceled) || errors.Is(ret, context.DeadlineExceeded) {
		ret = nil
	}
	return ret
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
		for i, c := range e.collectors {
			wg.Go(func() {
				next := make(map[string]struct{})

				startTime := time.Now()
				r, err := c.Collect(ctx)
				health[i] = (err == nil)
				dur[i] = time.Since(startTime)

				prefix := c.MetricsPrefix()
				if err != nil {
					slog.ErrorContext(ctx, "failed to collect metrics", slog.Any("error", err), slog.String("collector", prefix))
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
			})
		}
		wg.Wait()

		const collectorPrefix = "collector"
		for i, c := range e.collectors {
			labels := map[string]string{
				"collector": c.MetricsPrefix(),
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
