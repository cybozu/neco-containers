package registry

import (
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/ciliumid"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/mock"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/bpf"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

var collectors []exporter.Collector

func init() {
	collectors = []exporter.Collector{
		// scope: cluster
		ciliumid.NewCollector(),
		mock.NewCollector(),

		// scope: node
		bpf.NewCollector(),
	}
}

func All() []exporter.Collector {
	return collectors
}
