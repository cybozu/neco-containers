package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCephExecuterUpdate(t *testing.T) {
	testcases := map[string]struct {
		rule   rule
		expect map[string][]metricValue
	}{
		"happy path": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"echo", autoscale_status_json},
				metrics: map[string]metric{
					"pool_count": {
						metricType: prometheus.GaugeValue,
						help:       "pool count of `ceph osd pool autoscale-status` command",
						jqFilter:   "[{value: . | length, labels: []}]",
					},
					"actual_capacity_ratio": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   "[.[] | {value: .actual_capacity_ratio, labels: [.pool_name]}]",
						labelKeys:  []string{"pool_name"},
					},
				},
			},
			expect: map[string][]metricValue{
				"pool_count": {{value: 2.0, labelValues: []string{}}},
				"actual_capacity_ratio": {
					{value: 0.0, labelValues: []string{"device_health_metrics"}},
					{value: 1.0802450726274402, labelValues: []string{"ceph-ssd-block-pool"}},
				},
			},
		},
		"command execution failed": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"false"},
			},
			expect: map[string][]metricValue{},
		},
		"invalid jq filter": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"echo", autoscale_status_json},
				metrics: map[string]metric{
					"pool_count": {
						metricType: prometheus.GaugeValue,
						help:       "pool count of `ceph osd pool autoscale-status` command",
						jqFilter:   "[{value: . | length, labels: []}]",
					},
					"pg_autoscale_mode": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   "[.[] | {value: .pg_autoscale_mode, labels: [.pool_name]}]",
						labelKeys:  []string{"pool_name"},
					},
					"do_not_exist": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   "[.[] | {value: .do_not_exist, labels: [.pool_name]}]",
						labelKeys:  []string{"pool_name"},
					},
				},
			},
			expect: map[string][]metricValue{
				"pool_count": {{value: 2.0, labelValues: []string{}}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ce := newExecuter(&tc.rule)
			ce.update()

			assert.Equal(t, tc.expect, ce.metricValues)
		})
	}
}
