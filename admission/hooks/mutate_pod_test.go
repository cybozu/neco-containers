package hooks

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func createPod(po *v1.Pod) *v1.Pod {
	err := k8sClient.Create(testCtx, po)
	Expect(err).NotTo(HaveOccurred())

	ret := &v1.Pod{}
	err = k8sClient.Get(testCtx, types.NamespacedName{Name: po.GetName(), Namespace: po.GetNamespace()}, ret)
	Expect(err).NotTo(HaveOccurred())
	return ret
}

var _ = Describe("mutate Pod webhook", func() {
	It("should specify ephemeral-storage request and limit to container", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
  name: test-container
  namespace: default
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
`
		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		out := createPod(po)
		Expect(out.Spec.Containers[0].Resources.Requests).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Limits).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Requests[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("200Mi")))
		Expect(out.Spec.Containers[0].Resources.Limits[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("1Gi")))
	})

	It("should specify ephemeral-storage request and limit to initContainer", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
 name: test-init-container
 namespace: default
spec:
 containers:
 - name: ubuntu
   image: quay.io/cybozu/ubuntu
   command: ["pause"]
 initContainers:
 - name: init
   image: quay.io/cybozu/ubuntu
   command: ["pause"]
`
		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		out := createPod(po)
		Expect(out.Spec.Containers[0].Resources.Requests).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Limits).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Requests[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("200Mi")))
		Expect(out.Spec.Containers[0].Resources.Limits[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("1Gi")))
		Expect(out.Spec.InitContainers[0].Resources.Requests).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.InitContainers[0].Resources.Limits).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.InitContainers[0].Resources.Requests[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("200Mi")))
		Expect(out.Spec.InitContainers[0].Resources.Limits[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("1Gi")))
	})

	It("should overwrite ephemeral-storage request and limit to containers", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
 name: test-should-not-append
 namespace: default
spec:
 containers:
 - name: ubuntu
   image: quay.io/cybozu/ubuntu
   command: ["pause"]
   resources:
     requests:
       ephemeral-storage: "1Gi"
     limits:
       ephemeral-storage: "2Gi"
 initContainers:
 - name: init
   image: quay.io/cybozu/ubuntu
   command: ["pause"]
   resources:
     requests:
       ephemeral-storage: "100Mi"
     limits:
       ephemeral-storage: "100Mi"
`
		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		out := createPod(po)
		Expect(out.Spec.Containers[0].Resources.Requests).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Limits).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.Containers[0].Resources.Requests[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("200Mi")))
		Expect(out.Spec.Containers[0].Resources.Limits[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("1Gi")))
		Expect(out.Spec.InitContainers[0].Resources.Requests).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.InitContainers[0].Resources.Limits).Should(HaveKey(v1.ResourceEphemeralStorage))
		Expect(out.Spec.InitContainers[0].Resources.Requests[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("200Mi")))
		Expect(out.Spec.InitContainers[0].Resources.Limits[v1.ResourceEphemeralStorage]).Should(Equal(resource.MustParse("1Gi")))
	})
})
