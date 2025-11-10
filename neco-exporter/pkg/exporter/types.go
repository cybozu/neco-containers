package exporter

import "context"

type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type Collector interface {
	// Metrics names will be "neco_<Scope>_<MetricsPrefix>_<Metric.Name>{<Metric.Labels>}".
	//   Scope: specify through --scope parameter
	//   MetricsPrefix: specify in main.go
	Collect(ctx context.Context) ([]*Metric, error)
}
