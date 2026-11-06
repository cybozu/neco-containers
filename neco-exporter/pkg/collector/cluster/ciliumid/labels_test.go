package ciliumid

import (
	"maps"
	"testing"
)

func TestSecurityLabelsToPromLabels(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: map[string]string{},
		},
		{
			name: "typical identity",
			input: map[string]string{
				"k8s:identity.neco.cybozu.io/app":                                "compactor",
				"k8s:identity.neco.cybozu.io/name":                               "compactor",
				"k8s:io.cilium.k8s.namespace.labels.argocd.argoproj.io/instance": "tracing",
				"k8s:io.cilium.k8s.namespace.labels.team":                        "neco",
				"k8s:io.cilium.k8s.policy.cluster":                               "default",
				"k8s:io.kubernetes.pod.namespace":                                "tracing",
			},
			// the namespace security label is exposed as "namespace" instead
			expected: map[string]string{
				"label_k8s_identity_neco_cybozu_io_app":                                "compactor",
				"label_k8s_identity_neco_cybozu_io_name":                               "compactor",
				"label_k8s_io_cilium_k8s_namespace_labels_argocd_argoproj_io_instance": "tracing",
				"label_k8s_io_cilium_k8s_namespace_labels_team":                        "neco",
				"label_k8s_io_cilium_k8s_policy_cluster":                               "default",
			},
		},
		{
			// Kubernetes allows empty label values (e.g. `foo: ""`)
			name: "empty value",
			input: map[string]string{
				"k8s:foo": "",
			},
			expected: map[string]string{
				"label_k8s_foo": "",
			},
		},
		{
			name: "conflict",
			input: map[string]string{
				"k8s:a.b": "dot",
				"k8s:a_b": "underscore",
				"k8s:a/b": "slash",
				"k8s:c":   "no-conflict",
			},
			expected: map[string]string{
				"label_k8s_a_b": "<CONFLICTED>",
				"label_k8s_c":   "no-conflict",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := securityLabelsToPromLabels(tc.input)
			if !maps.Equal(actual, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
