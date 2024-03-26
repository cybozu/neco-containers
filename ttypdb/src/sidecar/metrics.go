package main

import (
	"math"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const metricsNamespace = "ttypdb_sidecar"

func initMetrics(logger *zap.Logger) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "ttys",
			Help:      "Number of controlling terminals observed",
		},
		func() float64 {
			ttys, err := ttyCount()
			if err != nil {
				logger.Error("failed to count ttys", zap.Error(err))
				return math.NaN()
			}
			return float64(ttys)
		},
	))
}
