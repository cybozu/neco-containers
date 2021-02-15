package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("prevent DELETE requests", func() {
	It("should allow to delete a PVC w/o the special annotation", func() {
		pvc := &corev1.PersistentVolumeClaim{}
		pvc.Name = "foo1"
		pvc.Namespace = "default"
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.DecimalSI),
		}
		scName := "local-storage"
		pvc.Spec.StorageClassName = &scName
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		err := k8sClient.Create(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny to delete a PVC w/ the special annotation", func() {
		pvc := &corev1.PersistentVolumeClaim{}
		pvc.Name = "foo2"
		pvc.Namespace = "default"
		pvc.Annotations = map[string]string{"admission.cybozu.com/prevent": "delete"}
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.DecimalSI),
		}
		scName := "local-storage"
		pvc.Spec.StorageClassName = &scName
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		err := k8sClient.Create(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, pvc)
		Expect(err).To(HaveOccurred())
	})
})
