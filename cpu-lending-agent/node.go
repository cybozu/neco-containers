package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// NodeAdvertiser publishes the current lending capacity of this node as an
// extended resource in the Node status (e.g.
// "cpu-lending.cybozu.io/lent-cpu: 2"), so that the scheduler places
// borrowers only onto nodes with lendable CPUs, bounds the borrower count per
// node, and `kubectl describe node` shows the lending headroom.
//
// The ledger follows reality, never the other way around: the caller updates
// cpusets first and advertises after. Shrinking the capacity does not evict
// running borrowers (Kubernetes checks inventory only at placement time);
// the resulting over-allocated ledger is harmless because new placements are
// blocked immediately and running borrowers already lost the lent CPUs.
type NodeAdvertiser struct {
	client       kubernetes.Interface
	nodeName     string
	resourceName string
}

// NewNodeAdvertiser creates a NodeAdvertiser. resourceName must be a valid
// extended resource name such as "cpu-lending.cybozu.io/lent-cpu".
func NewNodeAdvertiser(client kubernetes.Interface, nodeName, resourceName string) *NodeAdvertiser {
	return &NodeAdvertiser{client: client, nodeName: nodeName, resourceName: resourceName}
}

// Ensure makes the advertised capacity and allocatable equal to milliCPUs
// (the resource is denominated in milli-CPUs; 1000 per lent CPU).
// The scheduler admits pods against allocatable, so patching capacity alone
// is not enough: the kubelet does propagate extended resources from capacity
// to allocatable, but only on its own status sync, and a missing allocatable
// entry would otherwise never be repaired here. Ensure reads the current
// values and patches only on difference, so calling it on every reconcile is
// cheap and self-healing (a kubelet restart or node object recreation that
// drops the resource is repaired on the next reconcile).
func (n *NodeAdvertiser) Ensure(ctx context.Context, milliCPUs int) error {
	node, err := n.client.CoreV1().Nodes().Get(ctx, n.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", n.nodeName, err)
	}
	name := corev1.ResourceName(n.resourceName)
	capacity, capOK := node.Status.Capacity[name]
	allocatable, allocOK := node.Status.Allocatable[name]
	if capOK && allocOK && capacity.Value() == int64(milliCPUs) && allocatable.Value() == int64(milliCPUs) {
		return nil
	}

	// Strategic merge patch adds or replaces the single key and creates the
	// parent maps when absent (unlike a JSON patch "add"). The value is a
	// plain integer string: Quantity.String() would canonicalize 2000 to
	// "2k", which is a confusing display in `kubectl describe node`.
	value := strconv.Itoa(milliCPUs)
	patch, err := json.Marshal(map[string]any{
		"status": map[string]any{
			"capacity":    map[string]string{n.resourceName: value},
			"allocatable": map[string]string{n.resourceName: value},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal capacity patch: %w", err)
	}
	_, err = n.client.CoreV1().Nodes().Patch(ctx, n.nodeName,
		types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		return fmt.Errorf("failed to patch capacity of node %s: %w", n.nodeName, err)
	}
	return nil
}

// EnsureAnnotation makes the node annotation key equal to value, patching
// only on difference. It gives `kubectl describe node` a human-readable
// lending summary with the subtraction already done.
func (n *NodeAdvertiser) EnsureAnnotation(ctx context.Context, key, value string) error {
	node, err := n.client.CoreV1().Nodes().Get(ctx, n.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", n.nodeName, err)
	}
	if node.Annotations[key] == value {
		return nil
	}
	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]string{key: value},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal annotation patch: %w", err)
	}
	_, err = n.client.CoreV1().Nodes().Patch(ctx, n.nodeName,
		types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch annotation of node %s: %w", n.nodeName, err)
	}
	return nil
}
