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

func TestBuildMetricNamePanic(t *testing.T) {
	for _, v := range []string{`x\y`, `x"y`, "x\ny"} {
		t.Run(v, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for %q", v)
				}
			}()
			BuildMetricName("cluster", "test", "metric", map[string]string{"a": v})
		})
	}
}
