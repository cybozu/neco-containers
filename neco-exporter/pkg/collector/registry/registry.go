package registry

import (
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/ciliumid"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/mock"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/bpf"
)

var collectors []collector.Collector

func init() {
	collectors = []collector.Collector{
		// scope: cluster
		ciliumid.NewCollector(),
		mock.NewCollector(),

		// scope: node
		bpf.NewCollector(),
	}
}

func All() []collector.Collector {
	return collectors
}
