package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VictoriaMetrics/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConvertSquidCounter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	}))
	cases := []struct {
		name     string
		metric   []byte
		expected string
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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := io.NopCloser(bytes.NewReader(tc.metric))
			err := ConvertSquidCounter(body)
			assert.NoError(t, err)
			req := httptest.NewRequest("GET", "/metrics", nil)
			rec := httptest.NewRecorder()
			ts.Config.Handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expected)
		})
	}
}

func TestConvertSquidServiceTimes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	}))
	cases := []struct {
		name       string
		metric     []byte
		expected5  string
		expected60 string
	}{
		{
			name: "squid_service_times_http_requests_all_5_5min",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									HTTP Requests (All):   5%   0.00000  0.00000`),
			expected5:  "squid_service_times_http_requests_all_5percentile_5min 0",
			expected60: "squid_service_times_http_requests_all_5percentile_60min 0",
		},
		{
			name: "squid_service_times_cache_misses_10_5min",
			metric: []byte(`Service Time Percentiles            5 min    60 min:
									Cache Misses:          5%   0.00000  0.00000`),
			expected5:  "squid_service_times_cache_misses_5percentile_5min 0",
			expected60: "squid_service_times_cache_misses_5percentile_60min 0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := io.NopCloser(bytes.NewReader(tc.metric))
			err := ConvertSquidServiceTimes(body)
			assert.NoError(t, err)
			req := httptest.NewRequest("GET", "/metrics", nil)
			rec := httptest.NewRecorder()
			ts.Config.Handler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expected5)
			assert.Contains(t, rec.Body.String(), tc.expected60)

		})
	}
}
