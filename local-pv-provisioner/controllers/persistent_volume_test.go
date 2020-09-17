package controllers

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/google/go-cmp/cmp"
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
		pv := prepareLocalPV(ctx, "worker-1", false)

		Eventually(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			if apierrors.IsNotFound(err) {
				return nil
			}
			return errors.New("not deleted yet")
		}, 5*time.Second).Should(Succeed())
	})

	It("should not delete released PV without the label", func() {
		pv := prepareLocalPV(ctx, "worker-1", true)

		Consistently(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			return err
		}, 5*time.Second).Should(Succeed())

		err := k8sClient.Delete(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should not delete released PV with different node", func() {
		pv := prepareLocalPV(ctx, "worker-2", false)

		Consistently(func() error {
			var res corev1.PersistentVolume
			err := k8sClient.Get(ctx, types.NamespacedName{Name: pv.Name}, &res)
			return err
		}, 5*time.Second).Should(Succeed())

		err := k8sClient.Delete(ctx, &pv)
		Expect(err).ShouldNot(HaveOccurred())
	})
}

func prepareLocalPV(ctx context.Context, node string, witoutLabel bool) corev1.PersistentVolume {
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

	err := k8sClient.Create(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	pv.Status.Phase = corev1.VolumeReleased
	err = k8sClient.Status().Update(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	pv.ObjectMeta.Finalizers = []string{}
	err = k8sClient.Update(ctx, &pv)
	Expect(err).ShouldNot(HaveOccurred())

	return pv
}

func testFillDeleter() {
	It("should fill first specified bytes with zero", func() {
		tmpFile, _ := ioutil.TempFile("", "deleter")
		defer os.Remove(tmpFile.Name())
		err := exec.Command("dd", `if=/dev/urandom`, "of="+tmpFile.Name(), fmt.Sprintf("bs=%d", 1024), "count=11").Run()
		Expect(err).ShouldNot(HaveOccurred())

		deleter := &FillDeleter{
			FillBlockSize: 1024,
			FillCount:     10,
		}
		deleter.Delete(tmpFile.Name())

		zeroBlock := make([]byte, deleter.FillBlockSize)
		buffer := make([]byte, deleter.FillBlockSize)
		for i := uint(0); i < deleter.FillCount; i++ {
			_, err := tmpFile.Read(buffer)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cmp.Equal(buffer, zeroBlock)).Should(BeTrue())
		}

		_, err = tmpFile.Read(buffer)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cmp.Equal(buffer, zeroBlock)).Should(BeFalse())
	})
}
