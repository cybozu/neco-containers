package ciliumid

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/manager"
)

// clusterCollector reports CiliumIdentity resources in the cluster scope.
type clusterCollector struct {
	watcher *identityWatcher
}

var _ exporter.Collector = &clusterCollector{}

func NewClusterCollector() exporter.Collector {
	return &clusterCollector{
		watcher: newIdentityWatcher(),
	}
}

func (c *clusterCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *clusterCollector) MetricsPrefix() string {
	return metricsPrefix
}

func (c *clusterCollector) IsLeaderMetrics() bool {
	return true
}

func (c *clusterCollector) Setup(ctx context.Context) error {
	ctrl, err := manager.EnsureManager()
	if err != nil {
		return err
	}
	return c.watcher.setupWithManager(ctx, ctrl)
}

func (c *clusterCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
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

	for name, info := range c.watcher.getIdentities() {
		labels := securityLabelsToPromLabels(info.securityLabels)
		labels["identity"] = name
		labels["uid"] = info.uid
		labels["security_labels"] = formatSecurityLabels(info.securityLabels)
		if ns, ok := info.securityLabels[namespaceSecurityLabel]; ok {
			labels["namespace"] = ns
		}

		ret = append(ret, &exporter.Metric{
			Name:   "identity_info",
			Value:  1,
			Labels: labels,
		})
	}
	return ret, nil
}
