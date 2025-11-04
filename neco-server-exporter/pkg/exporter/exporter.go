package exporter

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type Exporter struct {
	running    atomic.Bool
	interval   time.Duration
	collectors []Collector

	log *slog.Logger
}

func NewExporter(interval time.Duration) *Exporter {
	return &Exporter{
		interval:   interval,
		collectors: make([]Collector, 0),
		log:        slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (e *Exporter) AddCollector(c Collector) error {
	if e.running.Load() {
		return errors.New("shouuld add collector before start")
	}
	e.collectors = append(e.collectors, c)
	return nil
}

func (e *Exporter) Start(ctx context.Context) error {
	go func() {
		if err := e.Run(ctx); err != nil {
			e.log.Error("failed to scrape metrics", slog.Any("error", err))
		}
	}()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	e.log.InfoContext(ctx, "start metrics server")

	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		e.log.Error("metrics server stopped", slog.Any("error", err))
		return err
	}
	return nil
}

func (e *Exporter) Run(ctx context.Context) error {
	const collectorSection = "collector"

	e.running.Store(true)
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

				section := c.SectionName()
				if err != nil {
					e.log.Error("failed to collect %s metrics: %w", section, err)
				} else {
					for _, m := range r {
						n := GetMetricsName(section, m.Name, m.Labels)
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

		for i, c := range e.collectors {
			labels := map[string]string{
				"collector": c.SectionName(),
			}

			n := GetMetricsName(collectorSection, "health", labels)
			counter := metrics.GetOrCreateFloatCounter(n)
			if health[i] {
				counter.Set(1)
			} else {
				counter.Set(0)
			}

			n = GetMetricsName(collectorSection, "process_seconds", labels)
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
