package exporter

import "context"

type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type Collector interface {
	// Metrics names will be "neco_server_<MetricsPrefix>_<MetricsName>{<MetricsLabels>}".
	MetricsPrefix() string
	Collect(ctx context.Context) ([]*Metric, error)
}
