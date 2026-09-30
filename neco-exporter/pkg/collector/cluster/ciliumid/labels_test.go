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
			expected: map[string]string{
				"label_k8s_identity_neco_cybozu_io_app":                                "compactor",
				"label_k8s_identity_neco_cybozu_io_name":                               "compactor",
				"label_k8s_io_cilium_k8s_namespace_labels_argocd_argoproj_io_instance": "tracing",
				"label_k8s_io_cilium_k8s_namespace_labels_team":                        "neco",
				"label_k8s_io_cilium_k8s_policy_cluster":                               "default",
				"label_k8s_io_kubernetes_pod_namespace":                                "tracing",
			},
		},
		{
			name: "non-k8s source",
			input: map[string]string{
				"reserved:host": "",
			},
			expected: map[string]string{
				"label_reserved_host": "",
			},
		},
		{
			name: "conflict",
			input: map[string]string{
				"k8s:a.b": "dot",
				"k8s:a_b": "underscore",
				"k8s:a/b": "slash",
			},
			// sorted order: "k8s:a.b" < "k8s:a/b" < "k8s:a_b"
			expected: map[string]string{
				"label_k8s_a_b":           "dot",
				"label_k8s_a_b_conflict1": "slash",
				"label_k8s_a_b_conflict2": "underscore",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// run multiple times to make sure the result does not depend on map iteration order
			for range 10 {
				actual := securityLabelsToPromLabels(tc.input)
				if !maps.Equal(actual, tc.expected) {
					t.Fatalf("expected %v, got %v", tc.expected, actual)
				}
			}
		})
	}
}
