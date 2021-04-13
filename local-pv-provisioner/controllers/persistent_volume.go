package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const watcherInterval = 1 * time.Hour

// PersistentVolumeReconciler reconciles a local PersistentVolume
type PersistentVolumeReconciler struct {
	client.Client
	NodeName string
	Deleter  Deleter
}

//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;delete

// Reconcile cleans up released local PV
func (r *PersistentVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pv corev1.PersistentVolume
	if err := r.Get(ctx, req.NamespacedName, &pv); err != nil {
		logger.Error(err, "unable to fetch PersistentVolume")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if name, ok := pv.ObjectMeta.Labels[localPVProvisionerLabelKey]; !ok || name != r.NodeName {
		return ctrl.Result{}, nil
	}

	if pv.Status.Phase != corev1.VolumeReleased {
		return ctrl.Result{}, nil
	}

	path := pv.Spec.Local.Path
	logger.Info("cleaning PersistentVolume", "path", path)
	if err := r.Deleter.Delete(path); err != nil {
		logger.Error(err, "unable to clean the device of PersistentVolume")
	}

	logger.Info("deleting PersistentVolume from api server")
	if err := r.Delete(context.Background(), &pv); err != nil {
		logger.Error(err, "unable to delete PersistentVolume")
		return ctrl.Result{}, err
	}

	logger.Info("successful to cleanup PersistentVolume")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager, nodeName string) error {
	ch := make(chan event.GenericEvent)
	watcher := &persistentVolumeWatcher{
		client:   mgr.GetClient(),
		ch:       ch,
		nodeName: nodeName,
		tick:     watcherInterval,
	}
	err := mgr.Add(watcher)
	if err != nil {
		return err
	}
	src := source.Channel{
		Source: ch,
	}

	pred := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		UpdateFunc:  func(event.UpdateEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return true },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolume{}).
		WithEventFilter(pred).
		Watches(&src, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

type persistentVolumeWatcher struct {
	client   client.Client
	ch       chan<- event.GenericEvent
	nodeName string
	tick     time.Duration
}

// Start implements Runnable.Start
func (w *persistentVolumeWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := w.fireEvent(context.Background())
			if err != nil {
				return err
			}
		}
	}
}

func (w *persistentVolumeWatcher) fireEvent(ctx context.Context) error {
	var pvs corev1.PersistentVolumeList
	err := w.client.List(ctx, &pvs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{localPVProvisionerLabelKey: w.nodeName}),
	})
	if err != nil {
		return err
	}

	for _, pv := range pvs.Items {
		if pv.Status.Phase != corev1.VolumeReleased {
			continue
		}
		w.ch <- event.GenericEvent{
			Object: &pv,
		}
	}
	return nil
}
