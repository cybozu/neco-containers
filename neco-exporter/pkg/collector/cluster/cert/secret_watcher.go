package cert

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

type secretWatcher struct {
	mu         sync.Mutex
	expiration map[types.NamespacedName]time.Time
}

func newSecretWatcher() *secretWatcher {
	return &secretWatcher{
		expiration: make(map[types.NamespacedName]time.Time),
	}
}

func (w *secretWatcher) update(ctx context.Context, s *corev1.Secret, deleted bool) {
	// !! DO NOT EXPOSE SECRET CONTENT TO LOG !!
	nn := types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
	if deleted {
		w.mu.Lock()
		defer w.mu.Unlock()
		delete(w.expiration, nn)
		return
	}

	data, ok := s.Data[corev1.TLSCertKey]
	if !ok {
		return
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		slog.WarnContext(ctx, "failed to read DER from Secret", slog.String("namespace", s.Namespace), slog.String("name", s.Name))
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse Certificate", slog.String("namespace", s.Namespace), slog.String("name", s.Name))
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.expiration[nn] = cert.NotAfter
}

func (w *secretWatcher) getCertificateExpiration() map[types.NamespacedName]time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return maps.Clone(w.expiration)
}

func (w *secretWatcher) setupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	informer, err := mgr.GetCache().GetInformer(ctx, &corev1.Secret{})
	if err != nil {
		return err
	}

	handler := func(o any, deleted bool) {
		s, ok := o.(*corev1.Secret)
		if !ok {
			slog.WarnContext(ctx, "unknown object returned from informer")
			return
		}
		if s.Type != corev1.SecretTypeTLS {
			return
		}
		managed := slices.ContainsFunc(s.OwnerReferences, func(o metav1.OwnerReference) bool {
			return o.APIVersion == "cert-manager.io/v1" && o.Kind == "Certificate"
		})
		if !managed {
			w.update(ctx, s, deleted)
		}
	}
	informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { handler(obj, false) },
		UpdateFunc: func(oldObj, newObj any) { handler(newObj, false) },
		DeleteFunc: func(obj any) { handler(obj, true) },
	}, time.Hour)
	return nil
}
