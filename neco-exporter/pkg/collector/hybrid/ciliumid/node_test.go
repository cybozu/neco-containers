package ciliumid

import (
	"maps"
	"testing"

	"github.com/cilium/cilium/api/v1/models"
)

func newEndpoint(podName string, id int64, labels ...string) *models.Endpoint {
	return &models.Endpoint{
		Status: &models.EndpointStatus{
			ExternalIdentifiers: &models.EndpointIdentifiers{K8sPodName: podName},
			Identity:            &models.Identity{ID: id, Labels: labels},
		},
	}
}

func TestCountIdentityPods(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.Endpoint
		expected map[string]int
	}{
		{
			name:     "no endpoints",
			input:    nil,
			expected: map[string]int{},
		},
		{
			name: "pods sharing an identity",
			input: []*models.Endpoint{
				// label order reported by Cilium does not matter
				newEndpoint("a-1", 1000, "k8s:io.kubernetes.pod.namespace=default", "k8s:app=a"),
				newEndpoint("a-2", 1000, "k8s:app=a", "k8s:io.kubernetes.pod.namespace=default"),
				newEndpoint("b-1", 1001, "k8s:app=b", "k8s:io.kubernetes.pod.namespace=default"),
			},
			expected: map[string]int{
				"k8s:app=a,k8s:io.kubernetes.pod.namespace=default": 2,
				"k8s:app=b,k8s:io.kubernetes.pod.namespace=default": 1,
			},
		},
		{
			name: "non-pod endpoints and reserved identities",
			input: []*models.Endpoint{
				newEndpoint("", 1, "reserved:host"),
				newEndpoint("", 4, "reserved:health"),
				newEndpoint("initializing", 5, "reserved:init"),
				newEndpoint("a-1", 1000, "k8s:app=a"),
			},
			expected: map[string]int{
				"k8s:app=a": 1,
			},
		},
		{
			name: "endpoints without identity or status",
			input: []*models.Endpoint{
				{},
				{Status: &models.EndpointStatus{ExternalIdentifiers: &models.EndpointIdentifiers{K8sPodName: "no-identity"}}},
				{Status: &models.EndpointStatus{Identity: &models.Identity{ID: 1000, Labels: []string{"k8s:app=a"}}}},
			},
			expected: map[string]int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := countIdentityPods(tc.input)
			if !maps.Equal(actual, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
