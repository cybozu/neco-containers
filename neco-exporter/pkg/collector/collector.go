package collector

import "context"

type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type Collector interface {
	// Metrics names will be "neco_<Scope>_<MetricsPrefix>_<Metric.Name>{<Metric.Labels>}".
	// Currently, scope should be either of cluster (constants.ScopeCluster) or node (constants.ScopeNode).
	Scope() string
	MetricsPrefix() string

	// Run necessary setup.
	// NOTE: This function is called one-by-one for multiple Collectors.
	Setup() error

	// Collect relevant metrices.
	// NOTE: this function is called simultaneously with other Collectors.
	Collect(ctx context.Context) ([]*Metric, error)
}
