package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var preventSubNamespaceDeletionValidatorConfig = &PreventSubNamespaceDeletionValidatorConfig{
	Resources: []NamespacedResourceGVK{
		{Group: "", Version: "v1", Kind: "Service"},
	},
}

var _ = Describe("PreventSubNamespaceDeletion validator", func() {
	It("should deny deletion of a namespace that has a configured resource", func() {
		ns := &corev1.Namespace{}
		ns.Name = "pnsv-1"
		ns.Annotations = map[string]string{
			"admission.cybozu.com/i-am-sure-to-delete": "pnsv-1",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "pnsv-1",
				Name:      "test-svc",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		err = k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).To(HaveOccurred())
	})

	It("should allow deletion of a namespace that has no configured resources", func() {
		ns := &corev1.Namespace{}
		ns.Name = "pnsv-2"
		ns.Annotations = map[string]string{
			"admission.cybozu.com/i-am-sure-to-delete": "pnsv-2",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow deletion of a namespace that only has resources not in config", func() {
		ns := &corev1.Namespace{}
		ns.Name = "pnsv-3"
		ns.Annotations = map[string]string{
			"admission.cybozu.com/i-am-sure-to-delete": "pnsv-3",
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "pnsv-3",
				Name:      "test-cm",
			},
		}
		err = k8sClient.Create(testCtx, cm)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})
})
