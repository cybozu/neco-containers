package controllers

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	releasedPVField = "status.phase"
	watcherInterval = 1 * time.Second
)

// Deleter clean up the block device
type Deleter interface {
	Delete(path string) error
}

// PersistentVolumeReconciler reconciles local PersistentVolume
type PersistentVolumeReconciler struct {
	client.Client
	Log      logr.Logger
	nodeName string
	deleter  Deleter
}

// +kubebuilder:rbac:groups="",resources=persistentvolume,verbs=get;list;watch;delete

// Reconcile cleans up released local PV
func (r *PersistentVolumeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("persistentvolume", req.NamespacedName)

	var pv corev1.PersistentVolume
	if err := r.Get(ctx, req.NamespacedName, &pv); err != nil {
		log.Error(err, "unable to fetch PersistentVolume", "name", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if name, ok := pv.ObjectMeta.Labels[localPVProvisionerLabelKey]; !ok || name != r.nodeName {
		return ctrl.Result{}, nil
	}

	if pv.Status.Phase != corev1.VolumeReleased {
		return ctrl.Result{}, nil
	}

	path := pv.Spec.Local.Path
	log.Info("cleaning PersistentVolume", "name", req.NamespacedName, "path", path)
	if err := r.deleter.Delete(path); err != nil {
		log.Error(err, "unable to clean PersistentVolume, will retry by periodical reconciliation", "name", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	log.Info("deleting PersistentVolume from api server", "name", req.NamespacedName)
	if err := r.Delete(context.Background(), &pv); err != nil {
		log.Error(err, "unable to delete PersistentVolume", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	log.Info("successful to cleanup PersistentVolume", "name", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up Reconciler with Manager.
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager, nodeName string) error {
	ch := make(chan event.GenericEvent)
	watcher := NewPersistentVolumeWatcher(mgr.GetClient(), ch, nodeName, watcherInterval)
	err := mgr.Add(watcher)
	if err != nil {
		return err
	}
	src := source.Channel{
		Source: ch,
	}

	pred := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return false },
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

// NewPersistentVolumeWatcher creates new persistentVolumeWatcher
func NewPersistentVolumeWatcher(client client.Client, ch chan<- event.GenericEvent, nodeName string, tick time.Duration) manager.Runnable {
	return &persistentVolumeWatcher{
		client:   client,
		ch:       ch,
		nodeName: nodeName,
		tick:     tick,
	}
}

// Start implements Runnable.Start
func (w *persistentVolumeWatcher) Start(ch <-chan struct{}) error {
	ticker := time.NewTicker(w.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ch:
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
			Meta: &metav1.ObjectMeta{
				Name: pv.Name,
			},
		}
	}
	return nil
}

// FillDeleter fills first 100MByte with '\0'
type FillDeleter struct {
	fillBlockSize uint
	fillCount     uint
}

// Delete implements Deleter's method.
func (d *FillDeleter) Delete(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	zeroBlock := make([]byte, d.fillBlockSize)
	for i := uint(0); i < d.fillCount; i++ {
		_, err = file.Write(zeroBlock)
		if err != nil {
			return err
		}
	}

	return nil
}
