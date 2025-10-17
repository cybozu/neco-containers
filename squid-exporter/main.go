package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/VictoriaMetrics/metrics"
)

func requestHandler(logger *slog.Logger, squidClient SquidClient) {
	// counters
	counters, err := squidClient.GetCounters()
	if err != nil {
		logger.Error("error getting squid counters", "err", err)
		return
	}
	err = ConvertSquidCounter(logger, counters)
	if err != nil {
		logger.Error("failed to convert squid counters", "err", err)
		return
	}

	// info
	info, err := squidClient.GetInfo()
	if err != nil {
		logger.Error("error getting squid info", "err", err)
		return
	}
	err = ConvertSquidInfo(logger, info)
	if err != nil {
		logger.Error("failed to convert squid info", "err", err)
		return
	}

	// service_times
	serviceTimes, err := squidClient.GetServiceTimes()
	if err != nil {
		logger.Error("error getting squid service_times", "err", err)
		return
	}
	err = ConvertSquidServiceTimes(logger, serviceTimes)
	if err != nil {
		logger.Error("failed to convert squid service_time", "err", err)
	}
	logger.Info("successfully converted squid metrics")
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config := NewConfig()
	squidClient := NewSquidClient(config)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		requestHandler(logger, squidClient)
		metrics.WritePrometheus(w, false)
	})
	fmt.Printf("Starting squid-exporter on port %d\n", config.MetricsPort)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.MetricsPort), nil)
}
