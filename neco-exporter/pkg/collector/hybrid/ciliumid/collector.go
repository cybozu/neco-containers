// Package ciliumid reports Cilium identities in both scopes.
//
//   - cluster: CiliumIdentity resources (identity_count, identity_info)
//   - node: identities of Pods on the node (identity_pods)
//
// Both scopes format security_labels with the same function, so that their metrics can be joined.
package ciliumid

import (
	"context"
	"fmt"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/option"
)

// scopedCollector is the implementation for a single scope.
type scopedCollector interface {
	isLeaderMetrics() bool
	setup(ctx context.Context) error
	collect(ctx context.Context) ([]*exporter.Metric, error)
}

type ciliumIDCollector struct {
	// impl is selected by the scope in Setup.
	impl scopedCollector
}

var _ exporter.Collector = &ciliumIDCollector{}

func NewCollector() exporter.Collector {
	return &ciliumIDCollector{}
}

func (c *ciliumIDCollector) Scope() string {
	return constants.ScopeHybrid
}

func (c *ciliumIDCollector) MetricsPrefix() string {
	return "ciliumid"
}

// IsLeaderMetrics must be called after Setup, as it depends on the scope.
func (c *ciliumIDCollector) IsLeaderMetrics() bool {
	return c.impl.isLeaderMetrics()
}

func (c *ciliumIDCollector) Setup(ctx context.Context) error {
	impl, err := newScopedCollector(option.Scope)
	if err != nil {
		return err
	}
	if err := impl.setup(ctx); err != nil {
		return err
	}
	c.impl = impl
	return nil
}

func (c *ciliumIDCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	return c.impl.collect(ctx)
}

func newScopedCollector(scope string) (scopedCollector, error) {
	switch scope {
	case constants.ScopeCluster:
		return newClusterCollector(), nil
	case constants.ScopeNode:
		return newNodeCollector(), nil
	default:
		return nil, fmt.Errorf("unsupported scope: %s", scope)
	}
}
