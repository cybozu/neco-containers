package ciliumid

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
)

type ciliumIDCollector struct {
}

var _ collector.Collector = &ciliumIDCollector{}

func NewCollector() collector.Collector {
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

func (c *ciliumIDCollector) Collect(ctx context.Context) ([]*collector.Metric, error) {
	return nil, nil
}
