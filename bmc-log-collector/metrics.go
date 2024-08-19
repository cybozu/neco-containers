package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

// var counterRequestFailed *prometheus.CounterVec
// var counterRequestSuccess *prometheus.CounterVec
//var counterRequestFailed prometheus.Counter
//var counterRequestSuccess prometheus.Counter

/*
var counterRequestFailed = promauto.NewCounter(

	prometheus.CounterOpts{
		Name: "failed_counter",
		Help: "The failed count for Redfish of BMC accessing",
	})

var counterRequestSuccess = promauto.NewCounter( /////////////<<<<  ここで競合

	prometheus.CounterOpts{
		Name: "success_counter",
		Help: "The success count for Redfish of BMC accessing",
	})
*/
var counterRequestFailed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "failed_counter",
		Help: "The failed count for Redfish of BMC accessing",
	},
	[]string{"status", "serial", "ip_addr"},
)

var counterRequestSuccess = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "success_counter",
		Help: "The success count for Redfish of BMC accessing",
	},
	[]string{"status", "serial", "ip_addr"},
)

func metrics() {

	reg := prometheus.NewRegistry()
	reg.MustRegister(counterRequestFailed)
	reg.MustRegister(counterRequestSuccess)

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	http.ListenAndServe(*addr, nil)
}
