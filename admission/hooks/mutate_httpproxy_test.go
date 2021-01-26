package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func testHTTPProxy(name string, annotations map[string]string) map[string]string {
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

	return ret.GetAnnotations()
}

var _ = Describe("mutate HTTPProxy webhook", func() {
	It("should have default annotation", func() {
		ann := testHTTPProxy("mhp1", map[string]string{})
		Expect(ann).To(HaveKeyWithValue(annotationKubernetesIngressClass, "secured"))
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
	})

	It("should not mutate annotations", func() {
		ann := testHTTPProxy("mhp2", map[string]string{annotationKubernetesIngressClass: "global"})
		Expect(ann).To(HaveKeyWithValue(annotationKubernetesIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
	})

	It("should not mutate annotations with projectcontour.io/ingress.class", func() {
		ann := testHTTPProxy("mhp3", map[string]string{annotationContourIngressClass: "global"})
		Expect(ann).To(HaveKeyWithValue(annotationContourIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationKubernetesIngressClass))
	})
})
