package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("SubNamespace deletion webhook", func() {
	It("should allow deletion when the target Namespace does not exist", func() {
		parent := &corev1.Namespace{}
		parent.Name = "snd-parent-1"
		Expect(k8sClient.Create(testCtx, parent)).To(Succeed())

		sn := newSubNamespaceUnstructured("snd-parent-1", "snd-nonexistent-child-1")
		Expect(k8sClient.Create(testCtx, sn)).To(Succeed())

		Expect(k8sClient.Delete(testCtx, sn)).To(Succeed())
	})

	It("should allow deletion when the target Namespace exists but has no resources", func() {
		parent := &corev1.Namespace{}
		parent.Name = "snd-parent-2"
		Expect(k8sClient.Create(testCtx, parent)).To(Succeed())

		child := &corev1.Namespace{}
		child.Name = "snd-child-2"
		Expect(k8sClient.Create(testCtx, child)).To(Succeed())

		sn := newSubNamespaceUnstructured("snd-parent-2", "snd-child-2")
		Expect(k8sClient.Create(testCtx, sn)).To(Succeed())

		Expect(k8sClient.Delete(testCtx, sn)).To(Succeed())
	})

	It("should deny deletion when the target Namespace has a resource", func() {
		parent := &corev1.Namespace{}
		parent.Name = "snd-parent-3"
		Expect(k8sClient.Create(testCtx, parent)).To(Succeed())

		child := &corev1.Namespace{}
		child.Name = "snd-child-3"
		Expect(k8sClient.Create(testCtx, child)).To(Succeed())

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "snd-child-3",
				Name:      "snd-blocker",
			},
		}
		Expect(k8sClient.Create(testCtx, cm)).To(Succeed())

		sn := newSubNamespaceUnstructured("snd-parent-3", "snd-child-3")
		Expect(k8sClient.Create(testCtx, sn)).To(Succeed())

		err := k8sClient.Delete(testCtx, sn)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`namespace "snd-child-3" still has resource`))
		Expect(err.Error()).To(ContainSubstring(`configmaps/snd-blocker`))
	})
})

func newSubNamespaceUnstructured(namespace, name string) *unstructured.Unstructured {
	sn := &unstructured.Unstructured{}
	sn.SetAPIVersion("accurate.cybozu.com/v2")
	sn.SetKind("SubNamespace")
	sn.SetNamespace(namespace)
	sn.SetName(name)
	return sn
}
