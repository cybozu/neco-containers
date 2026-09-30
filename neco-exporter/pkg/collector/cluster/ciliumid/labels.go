package ciliumid

import (
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

// conflictedLabelValue is set to a label when multiple security labels are converted to the same name.
// It is not a valid Kubernetes label value, so it cannot be confused with an actual value.
const conflictedLabelValue = "<CONFLICTED>"

// securityLabelsToPromLabels converts security labels into Prometheus labels.
// The namespace security label is skipped because it is exposed as the "namespace" label.
// If multiple keys are converted to the same name, the value of the label becomes conflictedLabelValue.
func securityLabelsToPromLabels(securityLabels map[string]string) map[string]string {
	labels := make(map[string]string, len(securityLabels))
	for k, v := range securityLabels {
		if k == namespaceSecurityLabel {
			continue
		}
		name := sanitizeLabelName(k)
		if _, ok := labels[name]; ok {
			v = conflictedLabelValue
		}
		labels[name] = v
	}
	return labels
}

// formatSecurityLabels formats security labels in the same way as Hubble metrics with
// the "identity" context, i.e. sorted "<source>:<key>=<value>" (or "<source>:<key>" if the
// value is empty) joined by commas.
func formatSecurityLabels(securityLabels map[string]string) string {
	entries := make([]string, 0, len(securityLabels))
	for k, v := range securityLabels {
		if v == "" {
			entries = append(entries, k)
		} else {
			entries = append(entries, k+"="+v)
		}
	}
	slices.Sort(entries)
	return strings.Join(entries, ",")
}
