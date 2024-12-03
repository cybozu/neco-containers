package main

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var counterRequestFailed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bmc_log_requests_failed_total",
		Help: "Failed count of accessing BMC to get the system event log",
	},
	[]string{"serial"},
)

var counterRequestSuccess = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bmc_log_requests_success_total",
		Help: "Succeeded count of accessing BMC to get the system event log",
	},
	[]string{"serial"},
)

func metrics(path string, port string) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(counterRequestFailed)
	reg.MustRegister(counterRequestSuccess)

	// Expose the registered metrics via HTTP.
	http.Handle(path, promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	slog.Error("error at ListenAndServe", "err", http.ListenAndServe(port, nil))
}

func deleteMetrics(serial string) {
	counterRequestSuccess.DeleteLabelValues(serial)
	counterRequestFailed.DeleteLabelValues(serial)
}
