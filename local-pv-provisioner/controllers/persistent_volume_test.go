package controllers

import (
	"context"
	"errors"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type deleterMock struct {
}

func (deleterMock) Delete(path string) error {
	return nil
}

func testPersistentVolumeReconciler() {
	ctx := context.Background()
	It("should delete released PV", func() {
		pv := prepareLocalPV("worker-1", corev1.VolumeReleased)
		err := k8sClient.Create(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
		pv.Status.Phase = corev1.VolumeReleased
		err = k8sClient.Status().Update(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.New("not deleted yet")
		}, 30*time.Second).Should(Succeed())
	})
}

func prepareLocalPV(node string, phase corev1.PersistentVolumePhase) corev1.PersistentVolume {
	return corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name: "local-pv",
			Labels: map[string]string{
				localPVProvisionerLabelKey: node,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/dev/crypt-disk/lun-0",
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse("1G"),
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      corev1.LabelHostname,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{node},
								},
							},
						},
					},
				},
			},
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: phase,
		},
	}
}
