package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	releasedPVField = "status.phase"
)

// NodeReconciler reconciles a Node object
type PersistentVolumeReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups="",resources=persistentvolume,verbs=get;list;watch;delete

// Reconcile cleans up released local PV
func (r *PersistentVolumeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("persistentvolume", req.NamespacedName)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up Reconciler with Manager.
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// err := mgr.GetFieldIndexer().IndexField(&corev1.PersistentVolume{}, releasedPVField, func(o runtime.Object) []string {
	// 	return []string{(o.(*corev1.PersistentVolume).Status.Phase)}
	// })
	// if err != nil {
	// 	return err
	// }

	pred := predicate.Funcs{
		UpdateFunc:  func(event.UpdateEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(pred).
		For(&corev1.PersistentVolume{}).
		Complete(r)
}
