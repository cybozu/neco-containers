package hooks

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var nonClusterIPTypes = []corev1.ServiceType{
	corev1.ServiceTypeLoadBalancer,
	corev1.ServiceTypeNodePort,
	corev1.ServiceTypeExternalName,
}

var _ = Describe("test for allowing rules for CREATE/UPDATE requests for Service", func() {
	It("should allow creating non-ClusterIP service with no externalIPs", func() {
		for i, t := range nonClusterIPTypes {
			svc := &corev1.Service{}
			svc.Name = fmt.Sprintf("allow1%d", i)
			svc.Namespace = "default"
			svc.Spec.Type = t
			svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
			svc.Spec.ExternalName = "foo"
			err := k8sClient.Create(testCtx, svc)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("should allow creating non-ClusterIP service with externalIPs", func() {
		for i, t := range nonClusterIPTypes {
			svc := &corev1.Service{}
			svc.Name = fmt.Sprintf("allow2%d", i)
			svc.Namespace = "default"
			svc.Spec.Type = t
			svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
			svc.Spec.ExternalIPs = []string{"10.0.0.1"}
			svc.Spec.ExternalName = "foo"
			err := k8sClient.Create(testCtx, svc)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("should allow creating ClusterIP service with no externalIPs", func() {
		svc := &corev1.Service{}
		svc.Name = "allow3"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow updating non-ClusterIP service with externalIPs", func() {
		for i, t := range nonClusterIPTypes {
			svc := &corev1.Service{}
			svc.Name = fmt.Sprintf("allow4%d", i)
			svc.Namespace = "default"
			svc.Spec.Type = t
			svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
			svc.Spec.ExternalName = "foo"
			err := k8sClient.Create(testCtx, svc)
			Expect(err).NotTo(HaveOccurred())

			svc.Spec.ExternalIPs = []string{"10.0.0.1"}
			err = k8sClient.Update(testCtx, svc)
			Expect(err).NotTo(HaveOccurred())
		}
	})
})

var _ = Describe("test for denying rules for CREATE/UPDATE requests for Service", func() {
	It("should deny creating ClusterIP service with externalIPs", func() {
		svc := &corev1.Service{}
		svc.Name = "deny1"
		svc.Namespace = "default"
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Ports = []corev1.ServicePort{{Name: "port", Port: 40000}}
		svc.Spec.ExternalIPs = []string{"10.0.0.1"}
		err := k8sClient.Create(testCtx, svc)
		Expect(err).To(HaveOccurred())
	})

	It("should deny updating ClusterIP service with externalIPs", func() {
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
