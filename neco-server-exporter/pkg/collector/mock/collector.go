package mock

import (
	"context"
	"errors"

	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/exporter"
)

// This Collector is used in CI.
// Don't use this in a real environment because it's useless.
type collector struct {
	count int
}

func NewCollector() exporter.Collector {
	return &collector{}
}

func (c *collector) MetricsPrefix() string {
	return "mock"
}

func (c *collector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	c.count++
	if c.count%2 == 0 {
		return nil, errors.New("test")
	}
	ret := []*exporter.Metric{
		{
			Name:  "test",
			Value: 100,
		},
	}
	return ret, nil
}
