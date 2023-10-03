package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var IpPolicyName = "restricted"
var httpproxyMutatorConfig = &HTTPProxyMutatorConfig{
	Policies: []HTTPProxyPolicy{
		{
			Name: IpPolicyName,
			IpAllowPolicy: []HTTPProxyIPFilterPolicy{
				{
					Source: "Peer",
					Cidr:   "192.0.2.0",
				},
			},
		},
	},
}

func testHTTPProxy(name string, annotations map[string]string, ingressClassNameField *string, ipAllowPolicy []HTTPProxyIPFilterPolicy) *unstructured.Unstructured {
	hp := &unstructured.Unstructured{}
	hp.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "projectcontour.io",
		Version: "v1",
		Kind:    "HTTPProxy",
	})
	hp.SetName(name)
	hp.SetNamespace("default")
	hp.SetAnnotations(annotations)
	hp.UnstructuredContent()["spec"] = map[string]interface{}{
		"routes": []map[string]interface{}{
			{
				"services": []interface{}{
					map[string]interface{}{
						"name": "dummy",
						"port": 80,
					},
					map[string]interface{}{
						"name": "dummy",
						"port": 443,
					},
				},
				"ipAllowPolicy": ipAllowPolicy,
			},
		},
	}
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
		hp := testHTTPProxy("mhp1", map[string]string{}, nil, nil)
		name, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		Expect(name).To(Equal("secured"))
	})

	It("should not mutate annotations", func() {
		hp := testHTTPProxy("mhp2", map[string]string{annotationKubernetesIngressClass: "global"}, nil, nil)
		_, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		ann := hp.GetAnnotations()
		Expect(ann).To(HaveKeyWithValue(annotationKubernetesIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("should not mutate annotations with projectcontour.io/ingress.class", func() {
		hp := testHTTPProxy("mhp3", map[string]string{annotationContourIngressClass: "global"}, nil, nil)
		_, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		ann := hp.GetAnnotations()
		Expect(ann).To(HaveKeyWithValue(annotationContourIngressClass, "global"))
		Expect(ann).ToNot(HaveKey(annotationKubernetesIngressClass))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("should not mutate .spec.ingressClassName field", func() {
		ingressClassName := "global"
		hp := testHTTPProxy("mhp4", map[string]string{}, &ingressClassName, nil)
		name, ok, err := unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		Expect(name).To(Equal("global"))
		ann := hp.GetAnnotations()
		Expect(ann).ToNot(HaveKey(annotationContourIngressClass))
		Expect(ann).ToNot(HaveKey(annotationKubernetesIngressClass))
	})

	It("should mutate .spec.routes[].ipAllowPolicy", func() {
		ipAllowPolicy := []HTTPProxyIPFilterPolicy{
			{
				Source: "Peer",
				Cidr:   "192.0.2.1",
			},
		}
		hp := testHTTPProxy("mhp5", map[string]string{annotationIpPolicy: IpPolicyName}, nil, ipAllowPolicy)
		routes, ok, err := unstructured.NestedSlice(hp.UnstructuredContent(), "spec", "routes")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		for _, route := range routes {
			policies, ok, err := unstructured.NestedFieldCopy(route.(map[string]interface{}), "ipAllowPolicy")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(Equal(true))
			Expect(policies).To(ContainElements(map[string]interface{}{
				"source": ipAllowPolicy[0].Source,
				"cidr":   ipAllowPolicy[0].Cidr,
			}))
		}
	})

	It("should not duplicate and mutate .spec.routes[].ipAllowPolicy, ", func() {
		hp := testHTTPProxy("mhp6", map[string]string{annotationIpPolicy: IpPolicyName}, nil, httpproxyMutatorConfig.Policies[0].IpAllowPolicy)
		routes, ok, err := unstructured.NestedSlice(hp.UnstructuredContent(), "spec", "routes")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		for _, route := range routes {
			policies, ok, err := unstructured.NestedFieldCopy(route.(map[string]interface{}), "ipAllowPolicy")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(Equal(true))
			Expect(policies).To(Equal([]interface{}{
				map[string]interface{}{
					"source": httpproxyMutatorConfig.Policies[0].IpAllowPolicy[0].Source,
					"cidr":   httpproxyMutatorConfig.Policies[0].IpAllowPolicy[0].Cidr,
				},
			}))
		}
	})

	It("should mutate .spec.routes[].ipAllowPolicy, if that field is patched to be removed", func() {
		hp := testHTTPProxy("mhp7", map[string]string{annotationIpPolicy: IpPolicyName}, nil, nil)
		routes, ok, err := unstructured.NestedSlice(hp.UnstructuredContent(), "spec", "routes")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		newRoutes := []interface{}{}
		for _, route := range routes {
			err := unstructured.SetNestedSlice(route.(map[string]interface{}), []interface{}{}, "ipAllowPolicy")
			Expect(err).ShouldNot(HaveOccurred())
			newRoutes = append(newRoutes, route)
		}
		err = unstructured.SetNestedSlice(hp.UnstructuredContent(), newRoutes, "spec", "routes")
		Expect(err).ShouldNot(HaveOccurred())

		err = k8sClient.Update(testCtx, hp)
		Expect(err).ShouldNot(HaveOccurred())

		ret := &unstructured.Unstructured{}
		ret.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "projectcontour.io",
			Version: "v1",
			Kind:    "HTTPProxy",
		})
		err = k8sClient.Get(testCtx, types.NamespacedName{Name: hp.GetName(), Namespace: hp.GetNamespace()}, ret)
		Expect(err).ShouldNot(HaveOccurred())

		routes, ok, err = unstructured.NestedSlice(ret.UnstructuredContent(), "spec", "routes")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ok).To(Equal(true))
		for _, route := range routes {
			policies, ok, err := unstructured.NestedFieldCopy(route.(map[string]interface{}), "ipAllowPolicy")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(Equal(true))
			Expect(policies).To(ContainElements(map[string]interface{}{
				"source": httpproxyMutatorConfig.Policies[0].IpAllowPolicy[0].Source,
				"cidr":   httpproxyMutatorConfig.Policies[0].IpAllowPolicy[0].Cidr,
			}))
		}
	})
})
