package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/VictoriaMetrics/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConvertSquidCounter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	}))
	cases := []struct {
		name        string
		metric      []byte
		expected    string
		notExpected string
	}{
		{
			name: "squid_counters_client_http.requests",
			metric: []byte(`sample_time = 1701938593.739082 (Thu, 07 Dec 2023 08:43:13 GMT)
							client_http.requests = 5`),
			expected: "squid_counters_client_http_requests_total 5",
		},
		{
			name: "cpu_time",
			metric: []byte(`sample_time = 1701938593.739082 (Thu, 07 Dec 2023 08:43:13 GMT)
							cpu_time = 59.389186`),
			expected: "cpu_time_total 59.389186",
		},
		{
			name: "has invallid_metric",
			metric: []byte(`sample_time = 1701938593.739082 (Thu, 07 Dec 2023 08:43:13 GMT)
							cpu_time ~ 59.389186`),
			notExpected: "cpu_time_total 59.389186",
		},
		{
			name: "has invallid and correct metric",
			metric: []byte(`sample_time = 1701938593.739082 (Thu, 07 Dec 2023 08:43:13 GMT)
							cpu_time ~ 59.389186
							client_http.requests = 5`),
			expected:    "squid_counters_client_http_requests_total 5",
			notExpected: "cpu_time_total",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metrics.UnregisterAllMetrics()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			body := io.NopCloser(bytes.NewReader(tc.metric))
			err := ConvertSquidCounter(logger, body)
			assert.NoError(t, err)
			req := httptest.NewRequest("GET", "/metrics", nil)
			rec := httptest.NewRecorder()
			ts.Config.Handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expected)
			if tc.notExpected != "" {
				assert.NotContains(t, rec.Body.String(), tc.notExpected)
			}
		})
	}
}

func TestConvertSquidServiceTimes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	}))
	cases := []struct {
		name          string
		metric        []byte
		expected5     string
		expected60    string
		notExpected5  string
		notExpected60 string
	}{
		{
			name: "squid_service_times_http_requests",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									HTTP Requests (All):   5%   1.00000  1.50000`),
			expected5:  `squid_service_times_http_requests_all{percentile="5", duration_minutes="5"} 1`,
			expected60: `squid_service_times_http_requests_all{percentile="5", duration_minutes="60"} 1.5`,
		},
		{
			name: "squid_service_times_cache_misses",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									Cache Misses:          5%   1.00000  1.50000`),
			expected5:  `squid_service_times_cache_misses{percentile="5", duration_minutes="5"} 1`,
			expected60: `squid_service_times_cache_misses{percentile="5", duration_minutes="60"} 1.5`,
		},
		{
			name: "has invallid_metric",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									Cache Misses          5%   1.00000  1.50000`),
			notExpected5:  `squid_service_times_cache_misses{percentile="5", duration_minutes="5"}`,
			notExpected60: `squid_service_times_cache_misses{percentile="5", duration_minutes="60"}`,
		},
		{
			name: "has invallid and correct metric",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									Cache Misses          5%   1.00000  1.50000
									Cache Hits:           5%   1.00000  1.50000`),
			expected5:     `squid_service_times_cache_hits{percentile="5", duration_minutes="5"} 1`,
			expected60:    `squid_service_times_cache_hits{percentile="5", duration_minutes="60"} 1.5`,
			notExpected5:  `squid_service_times_cache_misses{percentile="5", duration_minutes="5"}`,
			notExpected60: `squid_service_times_cache_misses{percentile="5", duration_minutes="60"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metrics.UnregisterAllMetrics()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			body := io.NopCloser(bytes.NewReader(tc.metric))
			err := ConvertSquidServiceTimes(logger, body)
			assert.NoError(t, err)
			req := httptest.NewRequest("GET", "/metrics", nil)
			rec := httptest.NewRecorder()
			ts.Config.Handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expected5)
			assert.Contains(t, rec.Body.String(), tc.expected60)
			if tc.notExpected5 != "" && tc.notExpected60 != "" {
				assert.NotContains(t, rec.Body.String(), tc.notExpected5)
				assert.NotContains(t, rec.Body.String(), tc.notExpected5)
			}
		})
	}
}
