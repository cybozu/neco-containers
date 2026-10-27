package ciliumid

import (
	"testing"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
)

func TestNewScopedCollector(t *testing.T) {
	testCases := []struct {
		scope         string
		expectError   bool
		leaderMetrics bool
	}{
		// cluster metrics are collected only by the leader, while node metrics are collected on every node
		{scope: constants.ScopeCluster, leaderMetrics: true},
		{scope: constants.ScopeNode, leaderMetrics: false},
		{scope: constants.ScopeHybrid, expectError: true},
		{scope: "unknown", expectError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.scope, func(t *testing.T) {
			impl, err := newScopedCollector(tc.scope)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error for scope %q", tc.scope)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if impl.isLeaderMetrics() != tc.leaderMetrics {
				t.Fatalf("expected isLeaderMetrics %v, got %v", tc.leaderMetrics, impl.isLeaderMetrics())
			}
		})
	}
}
