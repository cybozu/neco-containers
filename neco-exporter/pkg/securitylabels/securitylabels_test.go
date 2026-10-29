package securitylabels

import (
	"maps"
	"testing"
)

func TestFormat(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "",
		},
		{
			name: "typical identity",
			input: map[string]string{
				"k8s:io.kubernetes.pod.namespace":  "tracing",
				"k8s:identity.neco.cybozu.io/name": "compactor",
				"k8s:identity.neco.cybozu.io/app":  "compactor",
				"k8s:io.cilium.k8s.policy.cluster": "default",
			},
			expected: "k8s:identity.neco.cybozu.io/app=compactor,k8s:identity.neco.cybozu.io/name=compactor," +
				"k8s:io.cilium.k8s.policy.cluster=default,k8s:io.kubernetes.pod.namespace=tracing",
		},
		{
			// Kubernetes allows empty label values (e.g. `foo: ""`).
			// Same as Label.String() in Cilium, which Hubble uses, labels with empty values are
			// formatted as "<source>:<key>" without "=<value>", while the entries are still sorted.
			name: "empty value",
			input: map[string]string{
				"k8s:foo": "",
				"k8s:app": "bar",
				"k8s:zoo": "",
			},
			expected: "k8s:app=bar,k8s:foo,k8s:zoo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for range 10 {
				actual := Format(tc.input)
				if actual != tc.expected {
					t.Fatalf("expected %q, got %q", tc.expected, actual)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: map[string]string{},
		},
		{
			name:  "typical identity",
			input: []string{"k8s:io.kubernetes.pod.namespace=tracing", "k8s:app=compactor"},
			expected: map[string]string{
				"k8s:io.kubernetes.pod.namespace": "tracing",
				"k8s:app":                         "compactor",
			},
		},
		{
			// Label.String() in Cilium omits "=<value>" for an empty value
			name:     "empty value",
			input:    []string{"k8s:foo"},
			expected: map[string]string{"k8s:foo": ""},
		},
		{
			name:     "value containing =",
			input:    []string{"k8s:foo=a=b"},
			expected: map[string]string{"k8s:foo": "a=b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Parse(tc.input)
			if !maps.Equal(actual, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

// TestRoundTrip makes sure that labels reported by Cilium's endpoint API are formatted
// the same as the CiliumIdentity security-labels, i.e. the same as Label.String() joined by commas.
func TestRoundTrip(t *testing.T) {
	input := []string{"k8s:zoo", "k8s:app=bar", "k8s:io.kubernetes.pod.namespace=default"}
	expected := "k8s:app=bar,k8s:io.kubernetes.pod.namespace=default,k8s:zoo"
	if actual := Format(Parse(input)); actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
