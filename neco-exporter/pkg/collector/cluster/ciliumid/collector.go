package ciliumid

import (
	"context"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

type collector struct {
}

func NewCollector() (exporter.Collector, error) {
	// Leave it as a stub to the next PR
	return &collector{}, nil
}

func (c *collector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	return nil, nil
}
