package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var namespaceDeletionValidatorConfig = &NamespaceDeletionValidatorConfig{
	ProtectedResources: []NamespaceDeletionProtectedResource{
		{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		{Group: "", Version: "v1", Kind: "Service"},
	},
}

func newService(name, namespace string, annotations map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}
}

var _ = Describe("validate namespace deletion", func() {
	It("should allow to delete a namespace when Service has no prevent annotation", func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: "ns-deletion-1",
			Annotations: map[string]string{
				"admission.cybozu.com/i-am-sure-to-delete": "ns-deletion-1",
			},
		}}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		svc := newService("svc1", "ns-deletion-1", nil)
		err = k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny to delete a namespace when a Service has the prevent annotation", func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: "ns-deletion-2",
			Annotations: map[string]string{
				"admission.cybozu.com/i-am-sure-to-delete": "ns-deletion-2",
			},
		}}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		svc := newService("svc1", "ns-deletion-2", map[string]string{
			"admission.cybozu.com/prevent": "delete",
		})
		err = k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).To(HaveOccurred())
	})

	It("should allow to delete a namespace after removing the prevent annotation from a Service", func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: "ns-deletion-3",
			Annotations: map[string]string{
				"admission.cybozu.com/i-am-sure-to-delete": "ns-deletion-3",
			},
		}}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		svc := newService("svc1", "ns-deletion-3", map[string]string{
			"admission.cybozu.com/prevent": "delete",
		})
		err = k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).To(HaveOccurred())

		svc.Annotations = nil
		err = k8sClient.Update(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow to delete a namespace when a non-configured resource has the prevent annotation", func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: "ns-deletion-4",
			Annotations: map[string]string{
				"admission.cybozu.com/i-am-sure-to-delete": "ns-deletion-4",
			},
		}}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm1",
				Namespace: "ns-deletion-4",
				Annotations: map[string]string{
					"admission.cybozu.com/prevent": "delete",
				},
			},
		}
		err = k8sClient.Create(testCtx, cm)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())
	})
})
