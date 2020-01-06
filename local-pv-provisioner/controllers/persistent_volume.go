package controllers

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	storageClass     = "local-storage"
	hostNameLabelKey = "kubernetes.io/hostname"
	pvOwnerKey       = ".metadata.controller"
)

// SetupWithManager makes search index of owner references.
func (dd *DeviceDetector) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&corev1.PersistentVolume{}, pvOwnerKey, detectOwner); err != nil {
		return err
	}
	return nil
}

func detectOwner(rawObj runtime.Object) []string {
	pv := rawObj.(*corev1.PersistentVolume)
	owner := metav1.GetControllerOf(pv)
	if owner == nil {
		return nil
	}
	if owner.Kind != "Node" {
		return nil
	}
	return []string{owner.Name}
}

func (dd *DeviceDetector) pvName(devName string) string {
	hasher := sha1.New()
	hasher.Write([]byte(devName))
	return strings.Join([]string{"local", dd.nodeName, hex.EncodeToString(hasher.Sum(nil))[:10]}, "-")
}

func (dd *DeviceDetector) createPV(ctx context.Context, dev Device, ownerRef *metav1.OwnerReference) error {
	pvMode := corev1.PersistentVolumeBlock
	log := dd.log.WithValues("node", dd.nodeName)
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: dd.pvName(dev.Path),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: *resource.NewQuantity(dev.CapacityBytes, resource.BinarySI),
			},
			VolumeMode:                    &pvMode,
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              storageClass,
			PersistentVolumeSource:        corev1.PersistentVolumeSource{Local: &corev1.LocalVolumeSource{Path: dev.Path}},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: hostNameLabelKey, Operator: "In", Values: []string{dd.nodeName}},
					}},
				}},
			},
		},
	}
	op, err := ctrl.CreateOrUpdate(ctx, dd.Client, pv, func() error {
		pv.SetOwnerReferences([]metav1.OwnerReference{*ownerRef})
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create PV")
	} else {
		log.Info("PV successfully created", "operation", op)
	}
	return nil
}
