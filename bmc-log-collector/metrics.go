package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var counterRequestFailed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "failed_counter",
		Help: "The failed count for Redfish of BMC accessing",
	},
	[]string{"serial", "ip_addr"},
)

var counterRequestSuccess = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "success_counter",
		Help: "The success count for Redfish of BMC accessing",
	},
	[]string{"serial", "ip_addr"},
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
	http.ListenAndServe(port, nil)
}

func deleteMetrics(serial string, nodeIp string) {
	counterRequestSuccess.DeleteLabelValues(serial, nodeIp)
	counterRequestFailed.DeleteLabelValues(serial, nodeIp)
}
