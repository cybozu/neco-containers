package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validate DELETE requests", func() {
	It("should deny to delete a namespace w/o the special annotation", func() {
		ns := &corev1.Namespace{}
		ns.Name = "foo1"
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).To(HaveOccurred())
	})

	It("should deny to delete a namespace w/o proper value for the special annotation", func() {
		ns := &corev1.Namespace{}
		ns.Name = "foo2"
		ns.Annotations = map[string]string{
			"admission.cybozu.com/i-am-sure-to-delete": "bad",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).To(HaveOccurred())
	})

	It("should allow to delete a namespace with proper special annotation", func() {
		ns := &corev1.Namespace{}
		ns.Name = "foo3"
		ns.Annotations = map[string]string{
			"admission.cybozu.com/i-am-sure-to-delete": "foo3",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny to delete a quota", func() {
		ns := &corev1.Namespace{}
		ns.Name = "foo4"
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		quota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo4",
				Name:      "quota1",
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
				Scopes: []corev1.ResourceQuotaScope{
					corev1.ResourceQuotaScopeNotBestEffort,
				},
			},
		}
		err = k8sClient.Create(testCtx, quota)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, quota)
		Expect(err).To(HaveOccurred())
	})

	It("should not deny to delete a quota in a dev namespace", func() {
		ns := &corev1.Namespace{}
		ns.Name = "foo5"
		ns.Labels = map[string]string{
			"development": "true",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		quota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo5",
				Name:      "quota1",
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU: resource.MustParse("1"),
				},
				Scopes: []corev1.ResourceQuotaScope{
					corev1.ResourceQuotaScopeNotBestEffort,
				},
			},
		}
		err = k8sClient.Create(testCtx, quota)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, quota)
		Expect(err).NotTo(HaveOccurred())
	})
})
