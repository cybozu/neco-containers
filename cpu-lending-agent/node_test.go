package main

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

const testResource = "cpu-lending.cybozu.io/preemptible-millicpu"

func newTestNode(capacity, allocatable map[string]string) *corev1.Node {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node1"},
		Status: corev1.NodeStatus{
			Capacity:    corev1.ResourceList{},
			Allocatable: corev1.ResourceList{},
		},
	}
	for k, v := range capacity {
		node.Status.Capacity[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	for k, v := range allocatable {
		node.Status.Allocatable[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	return node
}

func patchCount(client *fake.Clientset) int {
	count := 0
	for _, action := range client.Actions() {
		if _, ok := action.(k8stesting.PatchAction); ok {
			count++
		}
	}
	return count
}

func TestNodeAdvertiserEnsure(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		capacity    map[string]string
		allocatable map[string]string
		milliCPUs   int
		wantPatched bool
	}{
		{
			name:        "advertise new resource",
			capacity:    map[string]string{"cpu": "12"},
			milliCPUs:   2000,
			wantPatched: true,
		},
		{
			name:        "update on change",
			capacity:    map[string]string{"cpu": "12", testResource: "2000"},
			allocatable: map[string]string{testResource: "2000"},
			milliCPUs:   0,
			wantPatched: true,
		},
		{
			name:        "no patch when converged",
			capacity:    map[string]string{"cpu": "12", testResource: "2000"},
			allocatable: map[string]string{testResource: "2000"},
			milliCPUs:   2000,
			wantPatched: false,
		},
		{
			name:        "capacity converged but allocatable missing is repaired",
			capacity:    map[string]string{"cpu": "12", testResource: "2000"},
			milliCPUs:   2000,
			wantPatched: true,
		},
		{
			name:        "zero is written, not deleted",
			capacity:    map[string]string{"cpu": "12"},
			milliCPUs:   0,
			wantPatched: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewClientset(newTestNode(tc.capacity, tc.allocatable))
			adv := NewNodeAdvertiser(client, "node1", testResource)
			if err := adv.Ensure(context.Background(), tc.milliCPUs); err != nil {
				t.Fatal(err)
			}
			if got := patchCount(client) > 0; got != tc.wantPatched {
				t.Errorf("patched = %v, want %v", got, tc.wantPatched)
			}
		})
	}
}

func TestNodeAdvertiserEnsureMissingNode(t *testing.T) {
	t.Parallel()
	client := fake.NewClientset()
	adv := NewNodeAdvertiser(client, "node1", testResource)
	if err := adv.Ensure(context.Background(), 1); err == nil {
		t.Error("Ensure succeeded, want error")
	}
}
