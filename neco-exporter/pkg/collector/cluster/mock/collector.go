package mock

import (
	"context"
	"errors"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
)

// This Collector is used in CI.
// Don't use this in a real environment because it's useless.
type mockCollector struct {
	count int
}

var _ collector.Collector = &mockCollector{}

func NewCollector() collector.Collector {
	return &mockCollector{}
}

func (c *mockCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *mockCollector) MetricsPrefix() string {
	return "mock"
}

func (c *mockCollector) Setup() error {
	return nil
}

func (c *mockCollector) Collect(ctx context.Context) ([]*collector.Metric, error) {
	c.count++
	if c.count%2 == 0 {
		return nil, errors.New("test")
	}
	ret := []*collector.Metric{
		{
			Name:  "test",
			Value: 100,
		},
	}
	return ret, nil
}
