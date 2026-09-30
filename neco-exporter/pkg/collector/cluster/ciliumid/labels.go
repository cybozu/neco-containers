package ciliumid

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// sanitizeLabelName converts a Cilium security label key (e.g. "k8s:app.kubernetes.io/name")
// into a valid Prometheus label name (e.g. "label_k8s_app_kubernetes_io_name").
func sanitizeLabelName(key string) string {
	var b strings.Builder
	b.WriteString("label_")
	for _, r := range key {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// securityLabelsToPromLabels converts security labels into Prometheus labels.
// Keys are processed in sorted order, and a key whose sanitized name is already used
// gets a "_conflict<N>" suffix so that the result is deterministic.
func securityLabelsToPromLabels(securityLabels map[string]string) map[string]string {
	ret := make(map[string]string, len(securityLabels))
	for _, k := range slices.Sorted(maps.Keys(securityLabels)) {
		name := sanitizeLabelName(k)
		if _, ok := ret[name]; ok {
			for i := 1; ; i++ {
				candidate := fmt.Sprintf("%s_conflict%d", name, i)
				if _, ok := ret[candidate]; !ok {
					name = candidate
					break
				}
			}
		}
		ret[name] = securityLabels[k]
	}
	return ret
}
