package main

import "github.com/prometheus/client_golang/prometheus"

var _ prometheus.Collector = &cephCollector{}

type cephCollector struct {
	executer       *cephExecuter
	describe       map[string]*prometheus.Desc
	describeFailed *prometheus.Desc
}

func newCollector(executer *cephExecuter, namespace string) *cephCollector {
	cc := &cephCollector{
		executer: executer,
		describe: map[string]*prometheus.Desc{},
	}
	for name, metric := range cc.executer.rule.metrics {
		cc.describe[name] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cc.executer.rule.name, name),
			metric.help, metric.labelKeys, nil)
	}
	cc.describeFailed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "failed_total"),
		"count of metrics export failure",
		[]string{"reason"}, prometheus.Labels{"subsystem": executer.rule.name})
	return cc
}

func (cc *cephCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range cc.describe {
		ch <- desc
	}
	ch <- cc.describeFailed
}

func (cc *cephCollector) Collect(ch chan<- prometheus.Metric) {
	cc.executer.mutex.RLock()
	defer cc.executer.mutex.RUnlock()
	for name, mVals := range cc.executer.metricValues {
		for _, mVal := range mVals {
			metric := cc.executer.rule.metrics[name]
			ch <- prometheus.MustNewConstMetric(
				cc.describe[name],
				metric.metricType,
				mVal.value,
				mVal.labelValues...,
			)
		}
	}
	for reason, count := range cc.executer.failedCounter {
		ch <- prometheus.MustNewConstMetric(
			cc.describeFailed,
			prometheus.CounterValue,
			float64(count),
			reason,
		)
	}
}
