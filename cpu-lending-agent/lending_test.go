package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// fakeTopology writes a sysfs-like tree where CPU 2n and 2n+1 are SMT
// siblings, and returns a Topology rooted there.
func fakeTopology(t *testing.T, numCPUs int) *Topology {
	t.Helper()
	root := t.TempDir()
	for cpu := 0; cpu < numCPUs; cpu++ {
		dir := filepath.Join(root, fmt.Sprintf("devices/system/cpu/cpu%d/topology", cpu))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		core := cpu / 2 * 2
		siblings := fmt.Sprintf("%d-%d\n", core, core+1)
		if err := os.WriteFile(filepath.Join(dir, "thread_siblings_list"), []byte(siblings), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return NewTopology(root)
}

func TestComputeLendable(t *testing.T) {
	t.Parallel()

	staticState := &PinnedState{
		Static:     true,
		SharedPool: mustParse(t, "0-1,6-11"),
		Reserved: map[string]map[string]CPUSet{
			"replica-uid": {"mysqld": mustParse(t, "2-3"), "sidecar": mustParse(t, "6")},
			"primary-uid": {"mysqld": mustParse(t, "4-5")},
		},
	}

	testCases := []struct {
		name           string
		st             *PinnedState
		lenders        []Lender
		container      string
		wholeCoresOnly bool
		want           string
	}{
		{
			name: "replica lends, primary does not",
			st:   staticState,
			lenders: []Lender{
				{UID: "replica-uid", Lending: true},
				{UID: "primary-uid", Lending: false},
			},
			container: "mysqld",
			want:      "2-3",
		},
		{
			name: "only the designated container donates CPUs",
			st:   staticState,
			lenders: []Lender{
				{UID: "replica-uid", Lending: true},
			},
			container: "mysqld",
			// The pinned "sidecar" container CPU 6 must not be lent.
			want: "2-3",
		},
		{
			name: "empty container name lends every container",
			st:   staticState,
			lenders: []Lender{
				{UID: "replica-uid", Lending: true},
			},
			want: "2-3,6",
		},
		{
			name:      "no lending lenders",
			st:        staticState,
			lenders:   []Lender{{UID: "primary-uid", Lending: false}},
			container: "mysqld",
			want:      "",
		},
		{
			name:      "lender without state entry contributes nothing",
			st:        staticState,
			lenders:   []Lender{{UID: "unknown-uid", Lending: true}},
			container: "mysqld",
			want:      "",
		},
		{
			name: "non-static state lends nothing (fail closed)",
			st:   &PinnedState{Static: false, Reserved: map[string]map[string]CPUSet{"replica-uid": {"mysqld": mustParse(t, "2-3")}}},
			lenders: []Lender{
				{UID: "replica-uid", Lending: true},
			},
			want: "",
		},
		{
			name: "whole cores only drops CPUs whose sibling belongs to others",
			st: &PinnedState{
				Static: true,
				// CPU 3's sibling is 2, which is neither lent nor in the
				// shared pool (someone else's exclusive CPU): 3 must not be
				// lent.
				Reserved: map[string]map[string]CPUSet{"replica-uid": {"mysqld": mustParse(t, "3,4-5")}},
			},
			lenders:        []Lender{{UID: "replica-uid", Lending: true}},
			container:      "mysqld",
			wholeCoresOnly: true,
			want:           "4-5",
		},
		{
			name: "whole cores only keeps CPUs whose sibling is in the shared pool",
			st: &PinnedState{
				Static: true,
				// CPU 3's sibling is 2, which is in the shared pool: the
				// borrower can use the whole core, so 3 may be lent.
				SharedPool: mustParse(t, "0-2"),
				Reserved:   map[string]map[string]CPUSet{"replica-uid": {"mysqld": mustParse(t, "3,4-5")}},
			},
			lenders:        []Lender{{UID: "replica-uid", Lending: true}},
			container:      "mysqld",
			wholeCoresOnly: true,
			want:           "3-5",
		},
		{
			name: "whole cores disabled keeps partial cores",
			st: &PinnedState{
				Static:   true,
				Reserved: map[string]map[string]CPUSet{"replica-uid": {"mysqld": mustParse(t, "3,4-5")}},
			},
			lenders:   []Lender{{UID: "replica-uid", Lending: true}},
			container: "mysqld",
			want:      "3-5",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ComputeLendable(tc.st, tc.lenders, tc.container, tc.wholeCoresOnly, fakeTopology(t, 12))
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tc.want {
				t.Errorf("ComputeLendable = %q, want %q", got.String(), tc.want)
			}
		})
	}
}

func TestDesiredBorrowerCPUSet(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		baseline string
		lendable string
		want     string
	}{
		{name: "lending", baseline: "0-1,6-11", lendable: "2-3", want: "0-3,6-11"},
		{name: "not lending is identity", baseline: "0-1,6-11", lendable: "", want: "0-1,6-11"},
		{name: "empty baseline", baseline: "", lendable: "2-3", want: "2-3"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DesiredBorrowerCPUSet(mustParse(t, tc.baseline), mustParse(t, tc.lendable))
			if got.String() != tc.want {
				t.Errorf("DesiredBorrowerCPUSet = %q, want %q", got.String(), tc.want)
			}
		})
	}
}

func TestRequestsResource(t *testing.T) {
	t.Parallel()
	makePod := func(requests map[string]string) *corev1.Pod {
		pod := &corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name: "c", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{}},
		}}}}
		for k, v := range requests {
			pod.Spec.Containers[0].Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
		return pod
	}
	testCases := []struct {
		name     string
		requests map[string]string
		want     bool
	}{
		{name: "requests the resource", requests: map[string]string{testResource: "1"}, want: true},
		{name: "no requests", requests: nil, want: false},
		{name: "other resource only", requests: map[string]string{"cpu": "100m"}, want: false},
		{name: "zero quantity", requests: map[string]string{testResource: "0"}, want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := RequestsResource(makePod(tc.requests), testResource); got != tc.want {
				t.Errorf("RequestsResource = %v, want %v", got, tc.want)
			}
		})
	}
}
