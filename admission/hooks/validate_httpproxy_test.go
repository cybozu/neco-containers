package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func fillHTTPProxy(name string, annotations map[string]string, ingressClassNameField *string) *unstructured.Unstructured {
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
	if ingressClassNameField != nil {
		unstructured.SetNestedField(hp.UnstructuredContent(), *ingressClassNameField, "spec", fieldIngressClassName)
	}
	return hp
}

var _ = Describe("validate HTTPProxy webhook with ", func() {
	It("should allow httpproxy with kubernetes.io/ingress.class", func() {
		hp := fillHTTPProxy("vhp1", map[string]string{annotationKubernetesIngressClass: "global"}, nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow httpproxy with projectcontour.io/ingress.class", func() {
		hp := fillHTTPProxy("vhp2", map[string]string{annotationContourIngressClass: "global"}, nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow httpproxy with .spec.ingressClassName", func() {
		ingressClassName := "global"
		hp := fillHTTPProxy("vhp6", map[string]string{}, &ingressClassName)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})

	if isHTTPProxyMutationDisabled() {
		// Mutation precedes validation.
		// Mutation sets default ingress class if not set by user.
		// So this test should not run if mutation is enabled.
		It("should deny httpproxy with no ingress class name", func() {
			hp := fillHTTPProxy("vhp8", nil, nil)
			err := k8sClient.Create(testCtx, hp)
			Expect(err).Should(HaveOccurred())
		})
	}

	It("should deny httpproxy to update kubernetes.io/ingress.class value", func() {
		hp := fillHTTPProxy("vhp3", map[string]string{annotationKubernetesIngressClass: "global"}, nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		ann := hp.GetAnnotations()
		ann[annotationKubernetesIngressClass] = "forest"
		hp.SetAnnotations(ann)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should deny httpproxy to update projectcontour.io/ingress.class value", func() {
		hp := fillHTTPProxy("vhp4", map[string]string{annotationContourIngressClass: "global"}, nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		ann := hp.GetAnnotations()
		ann[annotationContourIngressClass] = "forest"
		hp.SetAnnotations(ann)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should deny httpproxy to update .spec.ingressClassName value", func() {
		ingressClassName := "global"
		hp := fillHTTPProxy("vhp7", map[string]string{}, &ingressClassName)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		unstructured.SetNestedField(hp.UnstructuredContent(), "forest", "spec", fieldIngressClassName)
		err = k8sClient.Update(testCtx, hp)
		Expect(err).To(HaveOccurred())
	})

	It("should allow httpproxy to update other annotations", func() {
		hp := fillHTTPProxy("vhp5", map[string]string{annotationContourIngressClass: "global"}, nil)
		err := k8sClient.Create(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())

		hp.GetAnnotations()["foo"] = "forest"
		err = k8sClient.Update(testCtx, hp)
		Expect(err).NotTo(HaveOccurred())
	})
})
