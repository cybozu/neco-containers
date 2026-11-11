package registry

import (
	"testing"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
)

func TestCollectorsAreUniquePerScope(t *testing.T) {
	type key struct{ name, scope string }
	seen := make(map[key]bool)
	for _, c := range All() {
		if c.Scope() != constants.ScopeCluster && c.Scope() != constants.ScopeNode {
			t.Errorf("%s collector has invalid scope: %s", c.MetricsPrefix(), c.Scope())
		}
		k := key{c.MetricsPrefix(), c.Scope()}
		if seen[k] {
			t.Errorf("%s collector is registered twice in %s-scope", k.name, k.scope)
		}
		seen[k] = true
	}
}
