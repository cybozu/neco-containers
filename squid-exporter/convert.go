package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/VictoriaMetrics/metrics"
)

func ConvertSquidCounter(logger *slog.Logger, body io.ReadCloser) error {
	defer body.Close()
	scanner := bufio.NewScanner(body)
	scanner.Scan()
	for scanner.Scan() {
		metric := strings.Split(scanner.Text(), "=")
		if len(metric) != 2 {
			logger.Error("failed to parse squid counters")
			continue
		}
		metricName := strings.TrimSpace(strings.ReplaceAll(metric[0], ".", "_"))
		metricVal, err := strconv.ParseFloat(strings.TrimSpace(metric[1]), 64)
		if err != nil {
			return err
		}
		counter := metrics.GetOrCreateFloatCounter("squid_counters_" + metricName + "_total")
		counter.Set(metricVal)
	}
	return nil
}

func ConvertSquidServiceTimes(logger *slog.Logger, body io.ReadCloser) error {
	defer body.Close()
	scanner := bufio.NewScanner(body)
	scanner.Scan()
	r := strings.NewReplacer(
		" ", "_",
		"(", "",
		")", "",
		"-", "_",
	)
	for scanner.Scan() {
		metric := strings.Split(scanner.Text(), ":")
		if len(metric) != 2 {
			logger.Error("failed to parse squid service_times")
			continue
		}
		metricName := r.Replace(strings.ToLower(strings.TrimSpace(metric[0])))
		metricValues := strings.Split(strings.TrimLeft(metric[1], " "), "  ")
		if len(metricValues) != 3 {
			logger.Error("failed to parse squid service_times")
			continue
		}
		metricPercentile := strings.ReplaceAll(strings.TrimSpace(metricValues[0]), "%", "")
		metricVal5min, err := strconv.ParseFloat(strings.TrimSpace(metricValues[1]), 64)
		if err != nil {
			return err
		}
		metricVal60min, err := strconv.ParseFloat(strings.TrimSpace(metricValues[2]), 64)
		if err != nil {
			return err
		}
		counter := metrics.GetOrCreateFloatCounter(fmt.Sprintf(`squid_service_times_%s{percentile="%s", duration_minutes="%s"}`, metricName, metricPercentile, "5"))
		counter.Set(metricVal5min)
		counter = metrics.GetOrCreateFloatCounter(fmt.Sprintf(`squid_service_times_%s{percentile="%s", duration_minutes="%s"}`, metricName, metricPercentile, "60"))
		counter.Set(metricVal60min)
	}
	return nil
}
