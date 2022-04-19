package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

const autoscale_status_json = `
[
    {
        "actual_capacity_ratio": 1.2351049148057083e-05,
        "actual_raw_used": 1989198.0,
        "bias": 1.0,
        "bulk": false,
        "capacity_ratio": 1.2351049148057083e-05,
        "crush_root_id": -15,
        "effective_target_ratio": 0.0,
        "logical_used": 663066,
        "pg_autoscale_mode": "on",
        "pg_num_final": 1,
        "pg_num_ideal": 0,
        "pg_num_target": 1,
        "pool_id": 1,
        "pool_name": ".mgr",
        "raw_used": 1989198.0,
        "raw_used_rate": 3.0,
        "subtree_capacity": 161054982144,
        "target_bytes": 0,
        "target_ratio": 0.0,
        "would_adjust": false
    }
]
`

func TestCephExecuterUpdate(t *testing.T) {
	testcases := map[string]struct {
		rule   rule
		expect map[string]float64
	}{
		"happy path": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"echo", autoscale_status_json},
				metrics: map[string]metric{
					"pool_count": {
						metricType: prometheus.GaugeValue,
						help:       "pool count of `ceph osd pool autoscale-status` command",
						jqFilter:   ". | length",
					},
					"actual_capacity_ratio": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   ".[0].actual_capacity_ratio",
					},
				},
			},
			expect: map[string]float64{
				"pool_count":            1.0,
				"actual_capacity_ratio": 1.2351049148057083e-05,
			},
		},
		"command execution failed": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"false"},
			},
			expect: map[string]float64{},
		},
		"invalid jq filter": {
			rule: rule{
				name:    "osd_pool_autoscale_status",
				command: []string{"echo", autoscale_status_json},
				metrics: map[string]metric{
					"pool_count": {
						metricType: prometheus.GaugeValue,
						help:       "pool count of `ceph osd pool autoscale-status` command",
						jqFilter:   ". | length",
					},
					"pg_autoscale_mode": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   ".[0].pg_autoscale_mode",
					},
					"do_not_exist": {
						metricType: prometheus.GaugeValue,
						help:       "",
						jqFilter:   ".[0].do_not_exist",
					},
				},
			},
			expect: map[string]float64{
				"pool_count": 1.0,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ce := newExecuter(&tc.rule)
			ce.update()

			assert.Equal(t, tc.expect, ce.values)
		})
	}
}
