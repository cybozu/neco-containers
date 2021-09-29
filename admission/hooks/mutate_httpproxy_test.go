package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func testHTTPProxy(name string, annotations map[string]string, ingressClassNameField *string) *unstructured.Unstructured {
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

	err := k8sClient.Create(testCtx, hp)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	ret := &unstructured.Unstructured{}
	ret.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "projectcontour.io",
		Version: "v1",
		Kind:    "HTTPProxy",
	})
	err = k8sClient.Get(testCtx, types.NamespacedName{Name: name, Namespace: "default"}, ret)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return ret
}

var _ = Describe("mutate HTTPProxy webhook", func() {
	if isHTTPProxyMutationDisabled() {
		return
	}

	It("should have default ingress class name", func() {
		hp := testHTTPProxy("mhp1", map[string]string{}, nil)
		name, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		Expect(name).To(Equal("secured"))
	})

	It("should not mutate annotations", func() {
		hp := testHTTPProxy("mhp2", map[string]string{annotationKubernetesIngressClass: "global"}, nil)
		_, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		ann := hp.GetAnnotations()
		Expect(ann).To(HaveKeyWithValue(annotationKubernetesIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("should not mutate annotations with projectcontour.io/ingress.class", func() {
		hp := testHTTPProxy("mhp3", map[string]string{annotationContourIngressClass: "global"}, nil)
		_, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		ann := hp.GetAnnotations()
		Expect(ann).To(HaveKeyWithValue(annotationContourIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationKubernetesIngressClass))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("should not mutate .spec.ingressClassName field", func() {
		ingressClassName := "global"
		hp := testHTTPProxy("mhp4", map[string]string{}, &ingressClassName)
		name, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		Expect(name).To(Equal("global"))
		ann := hp.GetAnnotations()
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
		Expect(ann).ToNot(HaveKey(annotationKubernetesIngressClass))
	})
})
