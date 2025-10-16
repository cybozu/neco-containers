package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"regexp"
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

func ConvertSquidInfo(logger *slog.Logger, body io.ReadCloser) error {
	defer body.Close()

	fnFloat := func(name, value string) error {
		counter := metrics.GetOrCreateFloatCounter(name)
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		counter.Set(v)
		return nil
	}

	fnMinutes := func(name, value string) error {
		var v5, v60 float64
		_, err := fmt.Sscanf(value, "5min: %f%%, 60min: %f%%", &v5, &v60)
		if err != nil {
			return err
		}

		counterName := fmt.Sprintf(`%s{duration_minutes="%d"}`, name, 5)
		counter := metrics.GetOrCreateFloatCounter(counterName)
		counter.Set(v5 / 100)

		counterName = fmt.Sprintf(`%s{duration_minutes="%d"}`, name, 60)
		counter = metrics.GetOrCreateFloatCounter(counterName)
		counter.Set(v60 / 100)
		return nil
	}

	fnUsedFree := func(name, value string) error {
		// e.g. 0.0% used,  0.0% free
		v1, _, found := strings.Cut(value, ",")
		if !found {
			return fmt.Errorf("inappropriate format for %s: %s", name, value)
		}
		v1, found = strings.CutSuffix(strings.TrimSpace(v1), "% used")
		if !found {
			return fmt.Errorf("inappropriate format for %s: %s", name, value)
		}

		counter := metrics.GetOrCreateFloatCounter(name)
		v, err := strconv.ParseFloat(v1, 64)
		if err != nil {
			return err
		}
		counter.Set(v / 100)
		return nil
	}

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		title, value, found := strings.Cut(scanner.Text(), ":")
		title = strings.TrimSpace(title)
		value = strings.TrimSpace(value)
		fmt.Println(title)
		fmt.Println(value)
		if found {
			var err error
			switch title {
			// Cache information for squid:
			case "Hits as % of all requests":
				err = fnMinutes("squid_info_cache_hit_requests", value)
			case "Hits as % of bytes sent":
				err = fnMinutes("squid_info_cache_hit_bytes", value)
			case "Memory hits as % of hit requests":
				err = fnMinutes("squid_info_cache_memory_hit_requests", value)
			case "Disk hits as % of hit requests":
				err = fnMinutes("squid_info_cache_disk_hit_requests", value)
			case "Storage Swap capacity":
				err = fnUsedFree("squid_info_cache_disk_swap_capacity", value)
			case "Storage Mem capacity":
				err = fnUsedFree("squid_info_cache_memory_swap_capacity", value)

			// File descriptor usage for squid:
			case "Maximum number of file descriptors":
				err = fnFloat("squid_info_filefd_maximum", value)
			case "Largest file desc currently in use":
				err = fnFloat("squid_info_filefd_used_peak", value)
			case "Number of file desc currently in use":
				err = fnFloat("squid_info_filefd_used", value)
			case "Files queued for open":
				err = fnFloat("squid_info_filefd_queued", value)
			case "Available number of file descriptors":
				err = fnFloat("squid_info_filefd_available", value)
			case "Reserved number of file descriptors":
				err = fnFloat("squid_info_filefd_reserved", value)
			case "Store Disk files open":
				err = fnFloat("squid_info_filefd_store_disk", value)
			}
			if err != nil {
				return err
			}
		}
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
			logger.Error("failed to split squid service_times")
			continue
		}
		metricName := r.Replace(strings.ToLower(strings.TrimSpace(metric[0])))
		re := regexp.MustCompile(`\s+`)
		metricValues := re.Split(strings.TrimLeft(metric[1], " "), -1)
		if len(metricValues) != 3 {
			logger.Error("failed to parse squid service_times", "metric name", metricName, "values", metricValues)
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
