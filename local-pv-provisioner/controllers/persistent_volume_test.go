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
	if path == "/dev/crypt-disk/lun-0-broken" {
		return errors.New("broken device")
	}
	return nil
}

func testPersistentVolumeReconciler() {
	ctx := context.Background()
	It("should delete released PV", func() {
		pv := prepareLocalPV(ctx, "worker-1", false, false)

		Eventually(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.New("not deleted yet")
		}, 3*time.Second).Should(Succeed())
	})

	It("should not delete released PV without the label", func() {
		pv := prepareLocalPV(ctx, "worker-1", true, false)

		Consistently(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			return err
		}, 3*time.Second).Should(Succeed())

		err := k8sClient.Delete(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should not delete released PV with different node", func() {
		pv := prepareLocalPV(ctx, "worker-2", false, false)

		Consistently(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			return err
		}, 3*time.Second).Should(Succeed())

		err := k8sClient.Delete(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should delete released PV even with broken device", func() {
		pv := prepareLocalPV(ctx, "worker-1", false, true)

		Eventually(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.New("not deleted yet")
		}, 5*time.Second).Should(Succeed())
	})
}

func prepareLocalPV(ctx context.Context, node string, witoutLabel, broken bool) corev1.PersistentVolume {
	pv := corev1.PersistentVolume{
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
	}

	if witoutLabel {
		pv.ObjectMeta.Labels = nil
	}

	if broken {
		pv.Spec.Local.Path = "/dev/crypt-disk/lun-0-broken"
	}

	err := k8sClient.Create(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	pv.ObjectMeta.Finalizers = []string{}
	err = k8sClient.Update(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	pv.Status.Phase = corev1.VolumeReleased
	err = k8sClient.Status().Update(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	return pv
}
