package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/cybozu-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startServer(fetcher IBpfMapPressureFetcher, port uint) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector := newCollector(fetcher)
	prometheus.MustRegister(collector)
	go fetcher.Start(ctx)

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		_ = logger.Critical("failed to ListenAndServe", map[string]interface{}{log.FnError: err})
		return err
	}
	return nil
}

func main() {
	port := flag.Uint("port", 8080, "port number")
	configPath := flag.String("config", "/etc/bpf-map-pressure-exporter/config.yaml", "config file path")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		_ = logger.Critical("failed to load config", map[string]interface{}{log.FnError: err})
		os.Exit(1)
	}
	fetcher := newFetcher(config.MapNames, config.FetchInterval)
	if err := startServer(fetcher, *port); err != nil {
		os.Exit(1)
	}
}
