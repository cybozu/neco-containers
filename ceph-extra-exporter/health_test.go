package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckerServeHTTP(t *testing.T) {
	const threshold = 10 * time.Second

	testcases := map[string]struct {
		elapsedTimes   []time.Duration
		expectedStatus int
		expectedInBody string
	}{
		"all workers have updated recently": {
			elapsedTimes:   []time.Duration{time.Second, 9 * time.Second},
			expectedStatus: http.StatusOK,
		},
		"a worker is stuck": {
			elapsedTimes:   []time.Duration{time.Second, 11 * time.Second},
			expectedStatus: http.StatusServiceUnavailable,
			expectedInBody: "rule1",
		},
		"no worker is enabled": {
			elapsedTimes:   []time.Duration{},
			expectedStatus: http.StatusOK,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			executers := []*cephExecuter{}
			for i, elapsed := range tc.elapsedTimes {
				r := &rule{name: fmt.Sprintf("rule%d", i)}
				ce := newExecuter(r, executionInterval, commandTimeout)
				ce.lastUpdateTime = time.Now().Add(-elapsed)
				executers = append(executers, ce)
			}
			hc := newHealthChecker(executers, threshold)

			rw := httptest.NewRecorder()
			hc.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/v1/health", nil))

			assert.Equal(t, tc.expectedStatus, rw.Code)
			if tc.expectedInBody != "" {
				assert.Contains(t, rw.Body.String(), tc.expectedInBody)
			}
		})
	}
}

func TestMinHealthCheckThreshold(t *testing.T) {
	testRules := []rule{
		{
			name:    "one_metric",
			metrics: map[string]metric{"a": {}},
		},
		{
			name:    "three_metrics",
			target:  ruleTargetRGW,
			metrics: map[string]metric{"a": {}, "b": {}, "c": {}},
		},
	}

	testcases := map[string]struct {
		options  exportOptions
		expected time.Duration
	}{
		"the rule with the most metrics is enabled": {
			options: exportOptions{rgwMetrics: true, rbdMetrics: true},
			// executionInterval + (1 command + 3 jq) * commandTimeout
			expected: 100*time.Second + 4*10*time.Second,
		},
		"the rule with the most metrics is disabled": {
			options: exportOptions{rgwMetrics: false, rbdMetrics: false},
			// executionInterval + (1 command + 1 jq) * commandTimeout
			expected: 100*time.Second + 2*10*time.Second,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			actual := minHealthCheckThreshold(testRules, tc.options, 100*time.Second, 10*time.Second)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
