package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCephExecuterUpdate(t *testing.T) {
	testcases := map[string]struct {
		rule                rule
		expectedMetricValue map[string][]metricValue
		expectedFailedCount map[string]int
	}{
		"happy path": {
			rule: rule{
				name:    "test",
				command: []string{"echo", `[{"key": "key1", "value": 1}, {"key": "key2", "value": 2}]`},
				metrics: map[string]metric{
					"sum": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "[{value: . | map(.value) | add, labels: []}]",
					},
					"value": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "[.[] | {value: .value, labels: [.key]}]",
						labelKeys:  []string{"key"},
					},
				},
			},
			expectedMetricValue: map[string][]metricValue{
				"sum": {{value: 3, labelValues: []string{}}},
				"value": {
					{value: 1, labelValues: []string{"key1"}},
					{value: 2, labelValues: []string{"key2"}},
				},
			},
			expectedFailedCount: map[string]int{},
		},
		"command execution failed": {
			rule: rule{
				name:    "test",
				command: []string{"false"},
			},
			expectedMetricValue: map[string][]metricValue{},
			expectedFailedCount: map[string]int{
				"command": 1,
			},
		},
		"invalid jq filter": {
			rule: rule{
				name:    "test",
				command: []string{"echo", `[{"key": "key1", "value": 1}, {"key": "key2", "value": 2}]`},
				metrics: map[string]metric{
					"sum": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "[{value: . | map(.value) | add, labels: []}]",
					},
					"not_integer": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "[.[] | {value: .key, labels: [.key]}]",
						labelKeys:  []string{"key"},
					},
					"do_not_exist": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "[.[] | {value: .do_not_exist, labels: [.key]}]",
						labelKeys:  []string{"key"},
					},
					"broken_filter": {
						metricType: prometheus.GaugeValue,
						jqFilter:   "(",
						labelKeys:  []string{"key"},
					},
				},
			},
			expectedMetricValue: map[string][]metricValue{
				"sum": {{value: 3, labelValues: []string{}}},
			},
			expectedFailedCount: map[string]int{
				"jq":    1,
				"parse": 2,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ce := newExecuter(&tc.rule)
			ce.update()

			assert.Equal(t, tc.expectedMetricValue, ce.metricValues)
			assert.Subset(t, ce.failedCounter, tc.expectedFailedCount)
		})
	}
}
