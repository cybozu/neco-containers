package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
})
