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

	namespaceSecurityLabel = "k8s:io.kubernetes.pod.namespace"
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
	ns, ok, err := unstructured.NestedString(id.Object, "security-labels", namespaceSecurityLabel)
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

type identityInfo struct {
	uid            string
	securityLabels map[string]string
}

type identityWatcher struct {
	client client.Client

	mu            sync.Mutex
	identityCount map[string]int
	identities    map[string]identityInfo
}

func newIdentityWatcher() *identityWatcher {
	return &identityWatcher{
		identityCount: make(map[string]int),
		identities:    make(map[string]identityInfo),
	}
}

func (w *identityWatcher) update(ctx context.Context, id *unstructured.Unstructured, deleted bool) {
	ns, err := getIdentityNamespace(id)
	if err != nil {
		slog.WarnContext(ctx, "failed to get CiliumIdentity namespace")
		return
	}

	var info *identityInfo
	if !deleted {
		securityLabels, _, err := unstructured.NestedStringMap(id.Object, "security-labels")
		if err != nil {
			slog.WarnContext(ctx, "failed to get CiliumIdentity security labels", slog.Any("name", id.GetName()))
		} else {
			info = &identityInfo{
				uid:            string(id.GetUID()),
				securityLabels: securityLabels,
			}
		}
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

	if info != nil {
		w.identities[id.GetName()] = *info
	} else {
		delete(w.identities, id.GetName())
	}
}

func (w *identityWatcher) getNamespaceIdentityCount() map[string]int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return maps.Clone(w.identityCount)
}

func (w *identityWatcher) getIdentities() map[string]identityInfo {
	w.mu.Lock()
	defer w.mu.Unlock()

	return maps.Clone(w.identities)
}

func (w *identityWatcher) setupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()
	if err := indexer.IndexField(ctx, newCiliumIdentity(), indexKey, indexByNamespace); err != nil {
		return err
	}

	handler := func(o any, deleted bool) {
		if tombstone, ok := o.(cache.DeletedFinalStateUnknown); ok {
			o = tombstone.Obj
		}
		id, ok := o.(*unstructured.Unstructured)
		if !ok {
			slog.WarnContext(ctx, "unknown object returned from informer")
			return
		}
		w.update(ctx, id, deleted)
	}

	w.client = mgr.GetClient()
	informer, err := mgr.GetCache().GetInformer(ctx, newCiliumIdentity())
	if err != nil {
		return err
	}

	_, err = informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { handler(obj, false) },
		UpdateFunc: func(oldObj, newObj any) { handler(newObj, false) },
		DeleteFunc: func(obj any) { handler(obj, true) },
	}, time.Hour)
	return err
}
