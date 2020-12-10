package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("test for allowing rules for Service", func() {
	It("should allow creating Service with externalIPs empty", func() {
		svc := &corev1.Service{}
		svc.Name = "allow1"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow updating Service by changing port", func() {
		svc := &corev1.Service{}
		svc.Name = "allow2"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 50000}}
		err = k8sClient.Update(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("test for denying rules for Service", func() {
	It("should deny creating Service with externalIPs filled", func() {
		svc := &corev1.Service{}
		svc.Name = "deny1"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		svc.Spec.ExternalIPs = []string{"10.0.0.1"}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).To(HaveOccurred())
	})

	It("should deny updating Service by adding an IP to externalIPs", func() {
		svc := &corev1.Service{}
		svc.Name = "deny2"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())

		svc.Spec.ExternalIPs = []string{"10.0.0.1"}
		err = k8sClient.Update(testCtx, svc)
		Expect(err).To(HaveOccurred())
	})
})
