package registry

import (
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/cert"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/mock"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/cluster/networkfence"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/hybrid/ciliumid"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/bpf"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/cpupinning"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/kubeletconfig"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/collector/node/nicirq"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

var collectors []exporter.Collector

func init() {
	collectors = []exporter.Collector{
		// scope: cluster
		cert.NewCollector(),
		ciliumid.NewClusterCollector(),
		mock.NewCollector(),
		networkfence.NewCollector(),

		// scope: node
		bpf.NewCollector(),
		ciliumid.NewNodeCollector(),
		cpupinning.NewCollector(),
		kubeletconfig.NewCollector(),
		nicirq.NewCollector(),
	}
}

func All() []exporter.Collector {
	return collectors
}
