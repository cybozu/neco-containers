package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
)

func fillHTTPProxy(name string, annotations map[string]string) *contourv1.HTTPProxy {
	hp := &contourv1.HTTPProxy{}
	hp.Name = name
	hp.Namespace = "default"
	hp.Annotations = annotations
	return hp
}

var _ = Describe("validate HTTPProxy webhook with ", func() {
	It("should allow httpproxy with kubernetes.io/ingress.class", func() {
		hp := fillHTTPProxy("vhp1", map[string]string{annotationKubernetesIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow httpproxy with projectcontour.io/ingress.class", func() {
		hp := fillHTTPProxy("vhp2", map[string]string{annotationContourIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny updating httproxy with empty or no annotation", func() {
		hp := fillHTTPProxy("vhp3", map[string]string{annotationKubernetesIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.Annotations[annotationKubernetesIngressClass] = ""
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())

		delete(hp.Annotations, annotationKubernetesIngressClass)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})
})
