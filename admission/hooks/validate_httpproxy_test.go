package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func fillHTTPProxy(name string, annotations map[string]string) client.Object {
	hp := &unstructured.Unstructured{}
	hp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "projectcontour.io",
		Version: "v1",
		Kind:    "HTTPProxy",
	})
	hp.SetName(name)
	hp.SetNamespace("default")
	hp.SetAnnotations(annotations)
	hp.UnstructuredContent()["spec"] = map[string]interface{}{}
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

		ann := hp.GetAnnotations()
		ann[annotationKubernetesIngressClass] = "forest"
		hp.SetAnnotations(ann)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should deny httpproxy to update projectcontour.io/ingress.class value", func() {
		hp := fillHTTPProxy("vhp4", map[string]string{annotationContourIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		ann := hp.GetAnnotations()
		ann[annotationContourIngressClass] = "forest"
		hp.SetAnnotations(ann)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should allow httpproxy to update other annotations", func() {
		hp := fillHTTPProxy("vhp5", map[string]string{annotationContourIngressClass: "global"})
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.GetAnnotations()["foo"] = "forest"
		err = k8sClient.Update(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})
})
