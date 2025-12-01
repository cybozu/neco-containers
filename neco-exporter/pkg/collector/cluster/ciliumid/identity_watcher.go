package ciliumid

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	indexKey = "neco-exporter.ciliumid.namespace"
)

func newCiliumIdentity() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cilium.io",
		Version: "v2",
		Kind:    "CiliumIdentity",
	})
	return u
}

func newCiliumIdentityList() *unstructured.UnstructuredList {
	u := &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cilium.io",
		Version: "v2",
		Kind:    "CiliumIdentityList",
	})
	return u
}

func getIdentityNamespace(id *unstructured.Unstructured) (string, error) {
	ns, ok, err := unstructured.NestedString(id.Object, "security-labels", "k8s:io.kubernetes.pod.namespace")
	switch {
	case err != nil:
		return "", err
	case !ok:
		return "(null)", nil
	default:
		return ns, nil
	}
}

func indexByNamespace(obj client.Object) []string {
	id, ok := obj.(*unstructured.Unstructured)
	if !ok {
		slog.Warn("unknown object returned from informer", slog.Any("name", obj.GetName()))
		return nil
	}

	ns, err := getIdentityNamespace(id)
	if err != nil {
		slog.Warn("failed to get CiliumIdentity namespace", slog.Any("name", obj.GetName()))
		return nil
	}

	return []string{ns}
}

type identityWatcher struct {
	client client.Client

	mu            sync.Mutex
	identityCount map[string]int
}

func newIdentityWatcher() *identityWatcher {
	return &identityWatcher{
		identityCount: make(map[string]int),
	}
}

func (w *identityWatcher) update(ctx context.Context, id *unstructured.Unstructured) {
	ns, err := getIdentityNamespace(id)
	if err != nil {
		slog.WarnContext(ctx, "failed to get CiliumIdentity namespace")
		return
	}

	li := newCiliumIdentityList()
	if err := w.client.List(ctx, li, client.MatchingFields{indexKey: ns}); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to list by index: %v", err))
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(li.Items) > 0 {
		w.identityCount[ns] = len(li.Items)
	} else {
		delete(w.identityCount, ns)
	}
}

func (w *identityWatcher) getNamespaceIdentityCount() map[string]int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return maps.Clone(w.identityCount)
}

func (w *identityWatcher) setupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()
	if err := indexer.IndexField(ctx, newCiliumIdentity(), indexKey, indexByNamespace); err != nil {
		return err
	}

	handler := func(o any) {
		id, ok := o.(*unstructured.Unstructured)
		if !ok {
			slog.WarnContext(ctx, "unknown object returned from informer")
			return
		}
		w.update(ctx, id)
	}

	w.client = mgr.GetClient()
	informer, err := mgr.GetCache().GetInformer(ctx, newCiliumIdentity())
	if err != nil {
		return err
	}
	informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    handler,
		UpdateFunc: func(oldObj, newObj any) { handler(newObj) },
		DeleteFunc: handler,
	}, time.Hour)
	return nil
}
