package controllers

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	storageClass     = "local-storage"
	hostNameLabelKey = "kubernetes.io/hostname"
)

func (dd *DeviceDetector) pvName(devName string) string {
	hasher := sha1.New()
	hasher.Write([]byte(devName))
	return strings.Join([]string{"local", dd.nodeName, hex.EncodeToString(hasher.Sum(nil))[:10]}, "-")
}

func (dd *DeviceDetector) pvExists(pvList corev1.PersistentVolumeList, dev Device) bool {
	for _, pv := range pvList.Items {
		if pv.GetName() == dd.pvName(dev.Path) {
			return true
		}
	}
	return false
}

func (dd *DeviceDetector) createPV(ctx context.Context, dev Device, ownerRef *v1.OwnerReference) error {
	pvMode := corev1.PersistentVolumeBlock
	log := dd.log.WithValues("node", dd.nodeName)
	pv := &corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:   dd.pvName(dev.Path),
			Labels: map[string]string{nodeNameLabel: dd.nodeName},
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
		pv.SetOwnerReferences([]v1.OwnerReference{*ownerRef})
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create PV")
	} else {
		log.Info("PV successfully created", "operation", op)
	}
	return nil
}
