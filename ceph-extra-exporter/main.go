package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var rules = []rule{
	{
		name:    "osd_pool_autoscale_status",
		command: []string{"ceph", "osd", "pool", "autoscale-status", "-f", "json"},
		metrics: map[string]metric{
			"pool_count": {
				metricType: prometheus.GaugeValue,
				help:       "pool count of `ceph osd pool autoscale-status` command",
				jqFilter:   "[{value: . | length, labels: []}]",
			},
		},
	},
	{
		name:    "rgw_bucket_stats",
		command: []string{"radosgw-admin", "bucket", "stats"},
		metrics: map[string]metric{
			"s3_object_count": {
				metricType: prometheus.GaugeValue,
				help:       "s3 object count of `radosgw-admin bucket stats` command",
				jqFilter:   "[.[] | select(.usage.\"rgw.main\" != null) | {value: .usage.\"rgw.main\".num_objects, labels: [.bucket]}]",
				labelKeys:  []string{"bucket"},
			},
			"s3_size_bytes": {
				metricType: prometheus.GaugeValue,
				help:       "sum of s3 objects bytes `radosgw-admin bucket stats` command",
				jqFilter:   "[.[] | select(.usage.\"rgw.main\" != null) | {value: .usage.\"rgw.main\".size, labels: [.bucket]}]",
				labelKeys:  []string{"bucket"},
			},
			"s3_size_rounded_bytes": {
				metricType: prometheus.GaugeValue,
				help:       "sum of s3 objects bytes rounded to 4KBytes `radosgw-admin bucket stats` command",
				jqFilter:   "[.[] | select(.usage.\"rgw.main\" != null) | {value: .usage.\"rgw.main\".size_actual, labels: [.bucket]}]",
				labelKeys:  []string{"bucket"},
			},
		},
	},
	{
		name:    "osd_df",
		command: []string{"ceph", "osd", "df", "-f", "json"},
		metrics: map[string]metric{
			"crush_weight": {
				metricType: prometheus.GaugeValue,
				help:       "WEIGHT of `ceph osd df` command",
				jqFilter:   "[.nodes[] | {value: .crush_weight, labels: [.name]}]",
				labelKeys:  []string{"ceph_daemon"},
			},
		},
	},
}

//go:embed TAG
var version string

func init() {
	buildInfo := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "ceph_extra",
			Name:        "build_info",
			Help:        "Build info of the ceph-extra-exporter service.",
			ConstLabels: prometheus.Labels{"version": strings.TrimSpace(version)},
		}, func() float64 { return 1.0 })
	prometheus.MustRegister(buildInfo)
}

func startServer(rules []rule, port uint, doesRunRGWAdmin bool) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()
	for i := 0; i < len(rules); i++ {
		if !doesRunRGWAdmin && rules[i].command[0] == "radosgw-admin" {
			continue
		}
		wg.Add(1)
		go func(r *rule) {
			executer := newExecuter(r)
			prometheus.MustRegister(newCollector(executer, "ceph_extra"))
			executer.start(ctx)
			wg.Done()
		}(&rules[i])
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/health", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/v1/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("failed to ListenAndServe", "error", err)
		return err
	}

	return nil
}

func main() {
	port := flag.Uint("port", 8080, "port number")
	doesRunRGWAdmin := flag.Bool("export-rgw-metrics", true, "to export RGW related metrics or not")
	flag.Parse()
	if err := startServer(rules, *port, *doesRunRGWAdmin); err != nil {
		os.Exit(1)
	}
}
