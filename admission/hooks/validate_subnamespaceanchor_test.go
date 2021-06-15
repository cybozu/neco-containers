package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func fillSubnamespaceAnchor(name string) client.Object {
	subns := &unstructured.Unstructured{}
	subns.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "hnc.x-k8s.io",
		Version: "v1alpha2",
		Kind:    "SubnamespaceAnchor",
	})
	subns.SetName(name)
	subns.SetNamespace("default")
	subns.UnstructuredContent()["spec"] = map[string]interface{}{}
	return subns
}

var _ = Describe("validate SubnamespaceAnchor webhook with ", func() {
	It("should deny name not starting with dev-", func() {
		subns := fillSubnamespaceAnchor("test")
		err := k8sClient.Create(testCtx, subns)
		Expect(err).Should(HaveOccurred())
	})

	It("should allow name starting with dev-", func() {
		subns := fillSubnamespaceAnchor("dev-test")
		err := k8sClient.Create(testCtx, subns)
		Expect(err).ShouldNot(HaveOccurred())
	})
})
