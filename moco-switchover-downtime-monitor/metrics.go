package main

import "github.com/prometheus/client_golang/prometheus"

const metricsNamespace = "moco_switchover_downtime_monitor"

var downtimeBuckets = []float64{0, 0.1, 0.15, 0.25, 0.35, 0.5, 0.7, 1, 1.5, 2.5, 3.5, 5, 7, 10, 15, 25, 35, 50, 70, 100}
var downtimeLabels = []string{"cluster", "operation", "endpoint", "write"}

var downtimeGrossHistogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: metricsNamespace,
	Name:      "downtime_gross_seconds",
	Buckets:   downtimeBuckets,
}, downtimeLabels)
var downtimeNetHistogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: metricsNamespace,
	Name:      "downtime_net_seconds",
	Buckets:   downtimeBuckets,
}, downtimeLabels)

var checkFailureCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: metricsNamespace,
	Name:      "check_failure_total",
}, []string{"cluster", "operation", "reason"})

func init() {
	prometheus.MustRegister(downtimeGrossHistogramVec)
	prometheus.MustRegister(downtimeNetHistogramVec)
	prometheus.MustRegister(checkFailureCounter)
}
