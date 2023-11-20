package main

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type mockBpfMapPressureFetcher struct {
	metricsFunc func() []bpfMapPressure
	startFunc   func(context.Context)
}

func (f *mockBpfMapPressureFetcher) GetMetrics() []bpfMapPressure {
	return f.metricsFunc()
}

func (f *mockBpfMapPressureFetcher) Start(ctx context.Context) {
	f.startFunc(ctx)
}

func TestBpfMapPressureCollector(t *testing.T) {
	cases := []struct {
		name    string
		fetcher IBpfMapPressureFetcher
		expect  string
	}{
		{
			name: "success",
			fetcher: &mockBpfMapPressureFetcher{
				metricsFunc: func() []bpfMapPressure {
					return []bpfMapPressure{
						{
							mapId:       1,
							mapName:     "cilium_test_1",
							mapPressure: 0.1,
						},
						{
							mapId:       2,
							mapName:     "cilium_test_2",
							mapPressure: 0.2,
						},
					}
				},
			},
			expect: `# HELP bpf_map_pressure bpf map pressure
# TYPE bpf_map_pressure gauge
bpf_map_pressure{map_id="1",map_name="cilium_test_1"} 0.1
bpf_map_pressure{map_id="2",map_name="cilium_test_2"} 0.2
`,
		},
		{
			name: "duplicate maps",
			fetcher: &mockBpfMapPressureFetcher{
				metricsFunc: func() []bpfMapPressure {
					return []bpfMapPressure{
						{
							mapId:       1,
							mapName:     "cilium_test_1",
							mapPressure: 0.1,
						},
						{
							mapId:       1,
							mapName:     "cilium_test_2",
							mapPressure: 0.1,
						},
					}
				},
			},
			expect: `# HELP bpf_map_pressure bpf map pressure
# TYPE bpf_map_pressure gauge
bpf_map_pressure{map_id="1",map_name="cilium_test_1"} 0.1
bpf_map_pressure{map_id="1",map_name="cilium_test_1"} 0.1
`,
		},
		{
			name: "no maps",
			fetcher: &mockBpfMapPressureFetcher{
				metricsFunc: func() []bpfMapPressure {
					return []bpfMapPressure{}
				},
			},
			expect: `# HELP bpf_map_pressure bpf map pressure
# TYPE bpf_map_pressure gauge
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testutil.CollectAndCompare(newCollector(tc.fetcher), strings.NewReader(tc.expect), tc.name)
			assert.NoError(t, err)
		})
	}
}
