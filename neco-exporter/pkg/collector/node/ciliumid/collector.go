package ciliumid

import (
	"context"
	"fmt"
	"os"

	"github.com/cilium/cilium/api/v1/client/endpoint"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/client"
	"github.com/cilium/cilium/pkg/identity"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/securitylabels"
)

// nodeNameEnv is the env var the DaemonSet manifest is expected to populate from spec.nodeName.
const nodeNameEnv = "NODE_NAME"

type ciliumIDCollector struct {
	node         string
	ciliumClient *client.Client
}

var _ exporter.Collector = &ciliumIDCollector{}

func NewCollector() exporter.Collector {
	return &ciliumIDCollector{}
}

func (c *ciliumIDCollector) Scope() string {
	return constants.ScopeNode
}

func (c *ciliumIDCollector) MetricsPrefix() string {
	return "ciliumid"
}

func (c *ciliumIDCollector) IsLeaderMetrics() bool {
	return false
}

func (c *ciliumIDCollector) Setup(_ context.Context) error {
	node := os.Getenv(nodeNameEnv)
	if node == "" {
		return fmt.Errorf("%s environment variable is not set", nodeNameEnv)
	}

	cli, err := client.NewClient("")
	if err != nil {
		return fmt.Errorf("failed to open Cilium socket: %w", err)
	}

	c.node = node
	c.ciliumClient = cli
	return nil
}

func (c *ciliumIDCollector) Collect(_ context.Context) ([]*exporter.Metric, error) {
	response, err := c.ciliumClient.Endpoint.GetEndpoint(&endpoint.GetEndpointParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Cilium endpoints: %w", err)
	}

	counts := countIdentityPods(response.Payload)
	ret := make([]*exporter.Metric, 0, len(counts))
	for labels, count := range counts {
		ret = append(ret, &exporter.Metric{
			Name:  "identity_pods",
			Value: float64(count),
			Labels: map[string]string{
				"node":            c.node,
				"security_labels": labels,
			},
		})
	}
	return ret, nil
}

// countIdentityPods counts Pods by their security labels, formatted in the same way as
// security_labels of the cluster-scope ciliumid_identity_info metric.
//
// Endpoints that are not Pods (e.g. the host and cilium-health) and Pods with a reserved identity
// (e.g. reserved:init while waiting for their identity) are skipped, as they have no CiliumIdentity.
// Identities without Pods are not included, so their metrics disappear instead of becoming 0.
func countIdentityPods(endpoints []*models.Endpoint) map[string]int {
	ret := make(map[string]int)
	for _, ep := range endpoints {
		if ep.Status == nil || ep.Status.Identity == nil || ep.Status.ExternalIdentifiers == nil {
			continue
		}
		if ep.Status.ExternalIdentifiers.K8sPodName == "" {
			continue
		}
		if identity.NumericIdentity(ep.Status.Identity.ID).IsReservedIdentity() {
			continue
		}
		ret[securitylabels.Format(securitylabels.Parse(ep.Status.Identity.Labels))]++
	}
	return ret
}
