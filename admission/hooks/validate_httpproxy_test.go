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
	hp.Status.CurrentStatus = "dummy"
	hp.Status.Description = "dummy"
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

	It("should deny httpproxy with no ingress.class annotations", func() {
		hp := fillHTTPProxy("vhp2", nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).Should(HaveOccurred())

	})

	It("should deny httpproxy to update kubernetes.io/ingress.class value", func() {
		hp := fillHTTPProxy("vhp3", map[string]string{annotationKubernetesIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.Annotations[annotationKubernetesIngressClass] = "forest"
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should deny httpproxy to update projectcontour.io/ingress.class value", func() {
		hp := fillHTTPProxy("vhp4", map[string]string{annotationContourIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.Annotations[annotationContourIngressClass] = "forest"
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should allow httpproxy to update other annotations", func() {
		hp := fillHTTPProxy("vhp5", map[string]string{annotationContourIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.Annotations["foo"] = "forest"
		err = k8sClient.Update(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})
})
