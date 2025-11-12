package ciliumid

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

type ciliumIDCollector struct {
}

var _ exporter.Collector = &ciliumIDCollector{}

func NewCollector() exporter.Collector {
	return &ciliumIDCollector{}
}

func (c *ciliumIDCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *ciliumIDCollector) MetricsPrefix() string {
	return "ciliumid"
}

func (c *ciliumIDCollector) Setup() error {
	// TODO: setup shared informer to fetch CiliumIdentity resources
	return nil
}

func (c *ciliumIDCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	return nil, nil
}
