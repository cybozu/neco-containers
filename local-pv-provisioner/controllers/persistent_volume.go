package controllers

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// StorageClass is the name of StorageClass. It is set to pv.spec.storageClassName.
	StorageClass     = "local-storage"
	hostNameLabelKey = "kubernetes.io/hostname"
)

var (
	vpNameRegexp = regexp.MustCompile(`[^.0-9A-Za-z]+`)
)

func (dd *DeviceDetector) pvName(devPath string) string {
	tmp := strings.Join([]string{"local", dd.nodeName, filepath.Base(devPath)}, "-")
	return strings.ToLower(vpNameRegexp.ReplaceAllString(tmp, "-"))
}

func (dd *DeviceDetector) createPV(ctx context.Context, dev Device, node *corev1.Node) error {
	pvMode := corev1.PersistentVolumeBlock
	log := dd.log
	pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: dd.pvName(dev.Path)}}

	op, err := ctrl.CreateOrUpdate(ctx, dd.Client, pv, func() error {
		pv.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		pv.Spec.Capacity = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(dev.CapacityBytes, resource.BinarySI),
		}
		pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
			Required: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: hostNameLabelKey, Operator: "In", Values: []string{dd.nodeName}},
				}},
			}},
		}
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimRetain
		pv.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{Local: &corev1.LocalVolumeSource{Path: dev.Path}}
		pv.Spec.StorageClassName = StorageClass
		pv.Spec.VolumeMode = &pvMode
		return ctrl.SetControllerReference(node, pv, dd.scheme)
	})
	if err != nil {
		log.Error(err, "unable to create PV", "device", dev)
		return err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("PV successfully created", "operation", op, "device", dev)
	}
	return nil
}
