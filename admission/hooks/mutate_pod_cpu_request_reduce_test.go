package hooks

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func createAndGetPod(podManifest string) *corev1.Pod {
	po := &corev1.Pod{}
	err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096).Decode(po)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(context.Background(), po)
	Expect(err).NotTo(HaveOccurred())

	ret := &corev1.Pod{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: po.GetName(), Namespace: po.GetNamespace()}, ret)
	Expect(err).NotTo(HaveOccurred())
	return ret
}

func expectQuantitiesEqual(actual, expected resource.Quantity) {
	// NOTE:
	// We can not use `Equal()` to compare `resource.Quantity`.
	// For example, `resource.MustParse("1m")` and `resource.MustParse("0.001")` have the same value, but the internal structure is different.
	// So `Expect(resource.MustParse("1m")).To(Equal(resource.MustParse("0.001")))` causes the following error.
	//
	// [FAILED] Expected
	//   <resource.Quantity>: {
	//     i: {value: 1, scale: -3},
	//     d: {Dec: nil},
	//     s: "1m",
	//     Format: "DecimalSI",
	//   }
	// to equal
	//   <resource.Quantity>: {
	//     i: {value: 1, scale: -3},
	//     d: {Dec: nil},
	//     s: "",
	//     Format: "DecimalSI",
	//   }
	ExpectWithOffset(1, actual.Cmp(expected)).To(BeZero(), "actual=%s, expect=%s", actual.String(), expected.String())
}

var _ = Describe("CPU Request Reducer", func() {
	Describe("Unit Test", func() {
		It("should reduce quantity's value", func() {
			testcases := []struct {
				input    resource.Quantity
				expected resource.Quantity
			}{
				{
					resource.MustParse("0"),
					resource.MustParse("0"),
				},
				{
					resource.MustParse("0.001"),
					resource.MustParse("1m"),
				},
				{
					resource.MustParse("0.01"),
					resource.MustParse("5m"),
				},
				{
					resource.MustParse("0.1"),
					resource.MustParse("50m"),
				},
				{
					resource.MustParse("1"),
					resource.MustParse("500m"),
				},
				{
					resource.MustParse("64"),
					resource.MustParse("32000m"),
				},
				{
					resource.MustParse("0m"),
					resource.MustParse("0m"),
				},
				{
					resource.MustParse("1m"),
					resource.MustParse("1m"),
				},
				{
					resource.MustParse("2m"),
					resource.MustParse("1m"),
				},
				{
					resource.MustParse("3m"),
					resource.MustParse("1m"),
				},
				{
					resource.MustParse("4m"),
					resource.MustParse("2m"),
				},
				{
					resource.MustParse("4444m"),
					resource.MustParse("2222m"),
				},
			}

			for _, tt := range testcases {
				actual := reducedRequest(tt.input)
				expectQuantitiesEqual(actual, tt.expected)
			}
		})
	})

	Describe("Webhook", func() {
		It("should not reduce the CPU request of the pod with no CPU request", func() {
			podManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: reduce-test-pod-no-resources
  namespace: default
spec:
  initContainers:
  - name: init
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
`
			po := createAndGetPod(podManifest)
			Expect(po.Spec.InitContainers[0].Resources.Requests).NotTo(HaveKey(corev1.ResourceCPU))
			Expect(po.Spec.Containers[0].Resources.Requests).NotTo(HaveKey(corev1.ResourceCPU))
		})

		It("should not reduce the CPU request of the DaemonSet pod", func() {
			podManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: reduce-test-pod-daemonset
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: DaemonSet
    name: reduce-test
    uid: xxx
spec:
  initContainers:
  - name: init
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 0.5
        memory: 1Mi
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 3.5
        memory: 2Mi
  - name: ubuntu2
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        memory: 3Mi
`
			po := createAndGetPod(podManifest)
			expectQuantitiesEqual(po.Spec.InitContainers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("0.5"))
			expectQuantitiesEqual(po.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("3.5"))
			Expect(po.Spec.Containers[1].Resources.Requests).NotTo(HaveKey(corev1.ResourceCPU))
		})

		It("should not reduce the CPU request of the labeled pod", func() {
			podManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: reduce-test-pod-labeld
  namespace: default
  labels:
    admission.cybozu.com/prevent-cpu-request-reduce: "true"
spec:
  initContainers:
  - name: init
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 500m
        memory: 1Mi
  - name: init2
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      limits:
        cpu: 700m
        memory: 2Mi
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 1500m
        memory: 3Mi
`
			po := createAndGetPod(podManifest)
			expectQuantitiesEqual(po.Spec.InitContainers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("500m"))
			expectQuantitiesEqual(po.Spec.InitContainers[1].Resources.Requests[corev1.ResourceCPU], resource.MustParse("700m"))
			expectQuantitiesEqual(po.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("1500m"))
		})

		Describe("mutate", func() {
			podManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: reduce-test-pod-target
  namespace: default
spec:
  initContainers:
  - name: init
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 500m
        memory: 1Mi
  - name: init2
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      limits:
        cpu: 700m
  - name: init3
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 1.5
        memory: 3Mi
  - name: ubuntu2
    image: quay.io/cybozu/ubuntu
    command: ["pause"]
    resources:
      requests:
        cpu: 3
        memory: 3Mi
      limits:
        cpu: 4
        memory: 3Mi
`
			if isPodCPUReqeustReducerEnabled() {
				It("should reduce CPU requests when the webhook is enabled", func() {
					po := createAndGetPod(podManifest)
					expectQuantitiesEqual(po.Spec.InitContainers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("250m"))
					expectQuantitiesEqual(po.Spec.InitContainers[1].Resources.Requests[corev1.ResourceCPU], resource.MustParse("350m"))
					Expect(po.Spec.InitContainers[2].Resources.Requests).NotTo(HaveKey(corev1.ResourceCPU))
					expectQuantitiesEqual(po.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("750m"))
					expectQuantitiesEqual(po.Spec.Containers[1].Resources.Requests[corev1.ResourceCPU], resource.MustParse("1500m"))
				})
			}

			if !isPodCPUReqeustReducerEnabled() {
				It("should not reduce CPU requests when the webhook is disabled", func() {
					po := createAndGetPod(podManifest)
					expectQuantitiesEqual(po.Spec.InitContainers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("500m"))
					expectQuantitiesEqual(po.Spec.InitContainers[1].Resources.Requests[corev1.ResourceCPU], resource.MustParse("700m"))
					Expect(po.Spec.InitContainers[2].Resources.Requests).NotTo(HaveKey(corev1.ResourceCPU))
					expectQuantitiesEqual(po.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU], resource.MustParse("1500m"))
					expectQuantitiesEqual(po.Spec.Containers[1].Resources.Requests[corev1.ResourceCPU], resource.MustParse("3000m"))
				})
			}
		})
	})
})
