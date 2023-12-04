package main

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type bpfMapPressureCollector struct {
	describe *prometheus.Desc
	fetcher  IBpfMapPressureFetcher
}

func newCollector(fetcher IBpfMapPressureFetcher) *bpfMapPressureCollector {
	return &bpfMapPressureCollector{
		describe: prometheus.NewDesc(
			"bpf_map_pressure",
			"bpf map pressure",
			[]string{"map_id", "map_name"},
			nil,
		),
		fetcher: fetcher,
	}
}

func (c *bpfMapPressureCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.describe
}

func (c *bpfMapPressureCollector) Collect(ch chan<- prometheus.Metric) {
	for _, val := range c.fetcher.GetMetrics() {
		ch <- prometheus.MustNewConstMetric(
			c.describe,
			prometheus.GaugeValue,
			val.mapPressure,
			strconv.FormatUint(uint64(val.mapId), 10), val.mapName,
		)
	}
}
