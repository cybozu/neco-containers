package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"k8s.io/apimachinery/pkg/types"
)

func testHTTPProxy(name string, annotations map[string]string) *contourv1.HTTPProxy {
	hp := &contourv1.HTTPProxy{}
	hp.Name = name
	hp.Namespace = "default"
	hp.Annotations = annotations

	err := k8sClient.Create(testCtx, hp)
	Expect(err).NotTo(HaveOccurred())

	ret := &contourv1.HTTPProxy{}
	err = k8sClient.Get(testCtx, types.NamespacedName{Name: name, Namespace: "default"}, ret)
	Expect(err).NotTo(HaveOccurred())

	return ret
}

var _ = Describe("mutate HTTPProxy webhook", func() {
	It("should have default annotation", func() {
		hp := testHTTPProxy("mhp1", map[string]string{})
		Expect(hp.Annotations).To(HaveKeyWithValue(annotationKubernetesIngressClass, "secured"))
		Expect(hp.Annotations).ToNot(HaveKey(annotationContourIngressClass))
	})

	It("should not mutate annotations", func() {
		hp := testHTTPProxy("mhp2", map[string]string{annotationKubernetesIngressClass: "global"})
		Expect(hp.Annotations).To(HaveKeyWithValue(annotationKubernetesIngressClass, "global"))
		Expect(hp.Annotations).ToNot(HaveKey(annotationContourIngressClass))
	})

	It("should not mutate annotations with projectcontour.io/ingress.class", func() {
		hp := testHTTPProxy("mhp3", map[string]string{annotationContourIngressClass: "global"})
		Expect(hp.Annotations).To(HaveKeyWithValue(annotationContourIngressClass, "global"))
		Expect(hp.Annotations).ToNot(HaveKey(annotationKubernetesIngressClass))
	})
})
