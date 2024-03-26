package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "ttypdb_controller"

var metricsPollingSkipsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "polling_skips_total",
		Help:      "Number of polling skips",
	},
)

var metricsPollingDurationSecondsHistogram = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "polling_duration_seconds",
		Help:      "Polling duration took",
		Buckets:   []float64{0.001, 0.002, 0.004, 0.008, 0.012, 0.016, 0.024, 0.032, 0.064, 0.096, 0.128, 0.256, 0.512, 1.024},
	},
)

func init() {
	prometheus.MustRegister(
		metricsPollingSkipsCounter,
		metricsPollingDurationSecondsHistogram,
	)
}
