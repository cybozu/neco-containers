package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/VictoriaMetrics/metrics"
)

func ConvertSquidCounter(body io.ReadCloser) error {
	defer body.Close()
	metrics.UnregisterAllMetrics()
	scanner := bufio.NewScanner(body)
	scanner.Scan()
	for scanner.Scan() {
		metric := strings.Split(scanner.Text(), "=")
		metricName := strings.TrimSpace(strings.ReplaceAll(metric[0], ".", "_"))
		metricVal, err := strconv.ParseFloat(strings.TrimSpace(metric[1]), 64)
		if err != nil {
			return err
		}
		// fmt.Printf("%s:%v\n", metricName, metric[1])
		counter := metrics.GetOrCreateFloatCounter("squid_counters_" + metricName)
		counter.Set(metricVal)
	}
	return nil
}

func ConvertSquidServiceTimes(body io.ReadCloser) error {
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
		metricName := r.Replace(strings.ToLower(strings.TrimSpace(metric[0])))
		metricValues := strings.Split(strings.TrimLeft(metric[1], " "), "  ")
		metricPercentile := strings.ReplaceAll(strings.TrimSpace(metricValues[0]), "%", "")
		metricVal5min, err := strconv.ParseFloat(strings.TrimSpace(metricValues[1]), 64)
		if err != nil {
			return err
		}
		metricVal60min, err := strconv.ParseFloat(strings.TrimSpace(metricValues[2]), 64)
		if err != nil {
			return err
		}
		// fmt.Printf("%v,%v,%v\n", metricName, metricVal5min, metricVal60min)
		counter := metrics.GetOrCreateFloatCounter("squid_service_times_" + metricName + "_" + metricPercentile + "_5min")
		counter.Set(metricVal5min)
		counter = metrics.GetOrCreateFloatCounter("squid_service_times_" + metricName + "_" + metricPercentile + "_60min")
		counter.Set(metricVal60min)
	}
	return nil
}
