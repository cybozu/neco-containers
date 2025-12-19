package ciliumid

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/manager"
)

type ciliumIDCollector struct {
	watcher *identityWatcher
}

var _ exporter.Collector = &ciliumIDCollector{}

func NewCollector() exporter.Collector {
	return &ciliumIDCollector{
		watcher: newIdentityWatcher(),
	}
}

func (c *ciliumIDCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *ciliumIDCollector) MetricsPrefix() string {
	return "ciliumid"
}

func (c *ciliumIDCollector) IsLeaderMetrics() bool {
	return true
}

func (c *ciliumIDCollector) Setup(ctx context.Context) error {
	ctrl, err := manager.EnsureManager()
	if err != nil {
		return err
	}
	return c.watcher.setupWithManager(ctx, ctrl)
}

func (c *ciliumIDCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	m := c.watcher.getNamespaceIdentityCount()
	ret := make([]*exporter.Metric, 0, len(m))
	for k, v := range m {
		labels := map[string]string{
			"namespace": k,
		}
		nsMetric := &exporter.Metric{
			Name:   "identity_count",
			Value:  float64(v),
			Labels: labels,
		}
		ret = append(ret, nsMetric)
	}
	return ret, nil
}
