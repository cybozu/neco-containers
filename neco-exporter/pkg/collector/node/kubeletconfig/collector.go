package kubeletconfig

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

// configPath is the in-container path where kubelet's rendered configuration is expected.
// On a CKE-managed node the file actually lives at /etc/kubernetes/kubelet/config.yml; the
// DaemonSet's hostPath mount (neco-apps) is responsible for bind-mounting it to this path.
const configPath = "/var/lib/kubelet/config.yaml"

// nodeNameEnv is the env var the DaemonSet manifest is expected to populate from
// spec.nodeName, used to label metrics so they can be joined `on(node)` with metrics
// from other exporters (e.g. cAdvisor/cilium) that identify nodes the same way.
const nodeNameEnv = "NODE_NAME"

// systemReservedResources lists the systemReserved.* fields this collector reports,
// each labeled the same way kube_node_status_allocatable labels its resource/unit pairs.
var systemReservedResources = []struct {
	resource string
	unit     string
}{
	{resource: "cpu", unit: "core"},
	{resource: "memory", unit: "byte"},
}

// kubeletConfig is a minimal, ad-hoc subset of KubeletConfiguration; only the fields this
// collector needs are declared, rather than importing the full k8s.io/kubelet config API.
type kubeletConfig struct {
	SystemReserved map[string]string `json:"systemReserved"`
}

type reservation struct {
	resource string
	unit     string
	value    float64
}

type kubeletConfigCollector struct {
	node         string
	reservations []reservation
}

var _ exporter.Collector = &kubeletConfigCollector{}

func NewCollector() exporter.Collector {
	return &kubeletConfigCollector{}
}

func (c *kubeletConfigCollector) Scope() string {
	return constants.ScopeNode
}

func (c *kubeletConfigCollector) MetricsPrefix() string {
	return "kubelet"
}

func (c *kubeletConfigCollector) IsLeaderMetrics() bool {
	return false
}

func (c *kubeletConfigCollector) Setup(_ context.Context) error {
	node := os.Getenv(nodeNameEnv)
	if node == "" {
		return fmt.Errorf("%s environment variable is not set", nodeNameEnv)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read kubelet config %s: %w", configPath, err)
	}

	var cfg kubeletConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse kubelet config %s: %w", configPath, err)
	}

	reservations := make([]reservation, 0, len(systemReservedResources))
	for _, r := range systemReservedResources {
		raw, ok := cfg.SystemReserved[r.resource]
		if !ok {
			return fmt.Errorf("systemReserved.%s is not set in kubelet config %s", r.resource, configPath)
		}

		quantity, err := resource.ParseQuantity(raw)
		if err != nil {
			return fmt.Errorf("failed to parse systemReserved.%s %q: %w", r.resource, raw, err)
		}

		reservations = append(reservations, reservation{
			resource: r.resource,
			unit:     r.unit,
			value:    quantity.AsApproximateFloat64(),
		})
	}

	c.node = node
	c.reservations = reservations
	return nil
}

func (c *kubeletConfigCollector) Collect(_ context.Context) ([]*exporter.Metric, error) {
	ret := make([]*exporter.Metric, 0, len(c.reservations))
	for _, r := range c.reservations {
		ret = append(ret, &exporter.Metric{
			Name:  "system_reserved",
			Value: r.value,
			Labels: map[string]string{
				"node":     c.node,
				"resource": r.resource,
				"unit":     r.unit,
			},
		})
	}
	return ret, nil
}
