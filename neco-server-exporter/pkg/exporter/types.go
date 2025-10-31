package exporter

import "context"

type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type Collector interface {
	Name() string
	Collect(ctx context.Context) ([]*Metric, error)
}
