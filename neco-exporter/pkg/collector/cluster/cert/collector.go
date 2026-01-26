package cert

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/manager"
)

type certCollector struct {
	watcher *secretWatcher
}

var _ exporter.Collector = &certCollector{}

func NewCollector() exporter.Collector {
	return &certCollector{
		watcher: newSecretWatcher(),
	}
}

func (c *certCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *certCollector) MetricsPrefix() string {
	return "cert"
}

func (c *certCollector) IsLeaderMetrics() bool {
	return true
}

func (c *certCollector) Setup(ctx context.Context) error {
	ctrl, err := manager.EnsureManager()
	if err != nil {
		return err
	}
	return c.watcher.setupWithManager(ctx, ctrl)
}

func (c *certCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	m := c.watcher.getCertificateExpiration()
	ret := make([]*exporter.Metric, 0, len(m))
	for k, v := range m {
		labels := map[string]string{
			"namespace": k.Namespace,
			"name":      k.Name,
		}
		metric := &exporter.Metric{
			Name:   "expiration_timestamp_seconds",
			Value:  float64(v.UnixNano()),
			Labels: labels,
		}
		ret = append(ret, metric)
	}
	return ret, nil
}
