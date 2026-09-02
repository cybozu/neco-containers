package main

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Values of the log_type metrics label
const (
	metricLogTypeSel = "sel"
	metricLogTypeLc  = "lclog"
)

var counterRequestFailed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bmc_log_requests_failed_total",
		Help: "Failed count of accessing BMC to get the hardware log",
	},
	[]string{"serial", "log_type"},
)

var counterRequestSuccess = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bmc_log_requests_success_total",
		Help: "Succeeded count of accessing BMC to get the hardware log",
	},
	[]string{"serial", "log_type"},
)

var counterLcCatchupTruncated = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bmc_lclog_catchup_truncated_total",
		Help: "Count of the lifecycle log collection cycles that hit the page limit and skipped entries",
	},
	[]string{"serial"},
)

func metrics(path string, port string) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(counterRequestFailed)
	reg.MustRegister(counterRequestSuccess)
	reg.MustRegister(counterLcCatchupTruncated)

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
	counterRequestSuccess.DeletePartialMatch(prometheus.Labels{"serial": serial})
	counterRequestFailed.DeletePartialMatch(prometheus.Labels{"serial": serial})
	counterLcCatchupTruncated.DeleteLabelValues(serial)
}
