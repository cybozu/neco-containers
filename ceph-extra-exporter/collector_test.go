package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCephCollectorCollect(t *testing.T) {
	rule := rule{
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
	}
	ce := newExecuter(&rule)
	ce.update()
	cc := newCollector(ce, "ceph_extra")
	ch := make(chan prometheus.Metric)
	doneCh := make(chan struct{})
	go func() {
		cc.Collect(ch)
		doneCh <- struct{}{}
	}()
	metrics := []prometheus.Metric{}
	for {
		isDone := false
		select {
		case m := <-ch:
			metrics = append(metrics, m)
		case <-doneCh:
			isDone = true
		}
		if isDone {
			break
		}
	}

	// 5 metrics are:
	// ceph_extra_osd_pool_autoscale_status_pool_count
	// ceph_extra_osd_pool_autoscale_status_actual_capacity_ratio
	// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="command"}
	// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="jq"}
	// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="parse"}
	assert.Equal(t, 5, len(metrics))
}

func TestCephCollectorDescribe(t *testing.T) {
	rule := rule{
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
	}
	ce := newExecuter(&rule)
	ce.update()
	cc := newCollector(ce, "ceph_extra")
	ch := make(chan *prometheus.Desc)
	doneCh := make(chan struct{})
	go func() {
		cc.Describe(ch)
		doneCh <- struct{}{}
	}()
	descs := []*prometheus.Desc{}
	for {
		isDone := false
		select {
		case d := <-ch:
			descs = append(descs, d)
		case <-doneCh:
			isDone = true
		}
		if isDone {
			break
		}
	}

	// 3 describes are (* means a variable label):
	// ceph_extra_osd_pool_autoscale_status_pool_count
	// ceph_extra_osd_pool_autoscale_status_actual_capacity_ratio
	// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="*"}
	assert.Equal(t, 3, len(descs))
}
