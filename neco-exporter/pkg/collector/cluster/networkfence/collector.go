package networkfence

import (
	"context"
	"log/slog"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/manager"
)

func newNetworkFence() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "csiaddons.openshift.io",
		Version: "v1alpha1",
		Kind:    "NetworkFence",
	})
	return u
}

func newNetworkFenceList() *unstructured.UnstructuredList {
	u := &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "csiaddons.openshift.io",
		Version: "v1alpha1",
		Kind:    "NetworkFenceList",
	})
	return u
}

type networkFenceCollector struct {
	client client.Client
}

var _ exporter.Collector = &networkFenceCollector{}

func NewCollector() exporter.Collector {
	return &networkFenceCollector{}
}

func (c *networkFenceCollector) Scope() string {
	return constants.ScopeCluster
}

func (c *networkFenceCollector) MetricsPrefix() string {
	return "networkfence"
}

func (c *networkFenceCollector) IsLeaderMetrics() bool {
	return true
}

func (c *networkFenceCollector) Setup(ctx context.Context) error {
	mgr, err := manager.EnsureManager()
	if err != nil {
		return err
	}
	// Register the informer before mgr.Start() so that WaitForCacheSync covers it.
	obj := newNetworkFence()
	if _, err := mgr.GetCache().GetInformer(ctx, obj); err != nil {
		return err
	}
	c.client = mgr.GetClient()
	return nil
}

func (c *networkFenceCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	list := newNetworkFenceList()
	if err := c.client.List(ctx, list); err != nil {
		return nil, err
	}

	ret := make([]*exporter.Metric, 0, len(list.Items))
	for _, obj := range list.Items {
		name := obj.GetName()

		// spec.driver is optional in CRD validation.
		driver, _, err := unstructured.NestedString(obj.Object, "spec", "driver")
		if err != nil {
			slog.WarnContext(ctx, "spec.driver invalid in NetworkFence", slog.String("name", name), slog.Any("error", err))
			continue
		}

		fenceState, ok, err := unstructured.NestedString(obj.Object, "spec", "fenceState")
		if err != nil || !ok {
			slog.WarnContext(ctx, "spec.fenceState missing or invalid in NetworkFence", slog.String("name", name), slog.Any("error", err))
			continue
		}

		result, _, err := unstructured.NestedString(obj.Object, "status", "result")
		if err != nil {
			slog.WarnContext(ctx, "status.result invalid in NetworkFence", slog.String("name", name), slog.Any("error", err))
			continue
		}

		ret = append(ret, &exporter.Metric{
			Name: "info",
			Labels: map[string]string{
				"name":        name,
				"driver":      driver,
				"fence_state": fenceState,
				"result":      result,
			},
			Value: 1,
		})
	}

	return ret, nil
}
