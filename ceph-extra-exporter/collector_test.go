package main

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCephCollector(t *testing.T) {
	rule := rule{
		name:    "foo",
		command: []string{"echo", `{"value": 1}`},
		metrics: map[string]metric{
			"bar": {
				metricType: prometheus.GaugeValue,
				help:       "test metrics",
				jqFilter:   "[{value: .value, labels: []}]",
			},
		},
	}

	ce := newExecuter(&rule)
	ce.update()
	cc := newCollector(ce, "ceph_extra")

	testcases := []struct {
		name   string
		expect string
	}{
		{
			name: "ceph_extra_foo_bar",
			expect: `# HELP ceph_extra_foo_bar test metrics
# TYPE ceph_extra_foo_bar gauge
ceph_extra_foo_bar 1
`,
		},
		{
			name: "ceph_extra_failed_total",
			expect: `# HELP ceph_extra_failed_total count of metrics export failure
# TYPE ceph_extra_failed_total counter
ceph_extra_failed_total{reason="command",subsystem="foo"} 0
ceph_extra_failed_total{reason="jq",subsystem="foo"} 0
ceph_extra_failed_total{reason="parse",subsystem="foo"} 0
`,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CollectAndCompare(cc, strings.NewReader(tt.expect), tt.name)
			assert.NoError(t, err)
		})
	}
}
