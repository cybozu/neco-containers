package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCephCollectorCollect(t *testing.T) {
	testcases := []struct {
		rule          rule
		commandOutput string
		expect        int
	}{
		{
			rule:          rules[0],
			commandOutput: autoscale_status_json,
			// 4 metrics are:
			// ceph_extra_osd_pool_autoscale_status_pool_count
			// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="command"}
			// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="jq"}
			// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="parse"}
			expect: 4,
		},
		{
			rule:          rules[1],
			commandOutput: bucket_stats_json,
			// 12 metrics are:
			// ceph_extra_rgw_bucket_stats_s3_object_count{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"}
			// ceph_extra_rgw_bucket_stats_s3_object_count{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"}
			// ceph_extra_rgw_bucket_stats_s3_object_count{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"}
			// ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"}
			// ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"}
			// ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"}
			// ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"}
			// ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"}
			// ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"}
			// ceph_extra_failed_total{subsystem="rgw_bucket_stats", "reason"="command"}
			// ceph_extra_failed_total{subsystem="rgw_bucket_stats", "reason"="jq"}
			// ceph_extra_failed_total{subsystem="rgw_bucket_stats", "reason"="parse"}
			expect: 12,
		},
	}

	for _, tt := range testcases {
		rule := tt.rule
		rule.command = []string{"echo", tt.commandOutput}
		t.Run(rule.name, func(t *testing.T) {
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

			assert.Equal(t, tt.expect, len(metrics))
		})
	}
}

func TestCephCollectorDescribe(t *testing.T) {
	testcases := []struct {
		rule          rule
		commandOutput string
		expect        int
	}{
		{
			rule:          rules[0],
			commandOutput: autoscale_status_json,
			// 2 describes are (* means a variable label):
			// ceph_extra_osd_pool_autoscale_status_pool_count
			// ceph_extra_failed_total{subsystem="osd_pool_autoscale_status", "reason"="*"}
			expect: 2,
		},
		{
			rule:          rules[1],
			commandOutput: bucket_stats_json,
			// 4 describes are (* means a variable label):
			// ceph_extra_rgw_bucket_stats_s3_object_count{bucket="*"}
			// ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="*"}
			// ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="*"}
			// ceph_extra_failed_total{subsystem="rgw_bucket_stats", "reason"="*"}
			expect: 4,
		},
	}

	for _, tt := range testcases {
		rule := tt.rule
		rule.command = []string{"echo", tt.commandOutput}

		t.Run(rule.name, func(t *testing.T) {
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

			assert.Equal(t, tt.expect, len(descs))
		})
	}
}
