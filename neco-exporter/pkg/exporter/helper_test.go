package exporter

import "testing"

func TestBuildMetricName(t *testing.T) {
	testCases := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "nil labels",
			labels:   nil,
			expected: `neco_cluster_test_metric`,
		},
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: `neco_cluster_test_metric`,
		},
		{
			name:     "sorted labels",
			labels:   map[string]string{"b": "2", "a": "1"},
			expected: `neco_cluster_test_metric{a="1",b="2"}`,
		},
		{
			name:     "escaped values",
			labels:   map[string]string{"a": `x\y"z` + "\n"},
			expected: `neco_cluster_test_metric{a="x\\y\"z\n"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := BuildMetricName("cluster", "test", "metric", tc.labels)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
