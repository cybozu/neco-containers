package hooks

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("validate Pod webhook", func() {
	It("should allow a Pod having trusted images", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod-validate-1
  namespace: default
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
  - name: etcd
    image: quay.io/cybozu/etcd
  initContainers:
  - name: init1
    image: quay.io/cybozu/init
`

		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(testCtx, po)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny a Pod having untrustworthy container", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod-validate-2
  namespace: default
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
  - name: etcd
    image: mysql
  initContainers:
  - name: init1
    image: quay.io/cybozu/init
`

		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(testCtx, po)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny a Pod having untrustworthy initContainer", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod-validate-3
  namespace: default
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
  initContainers:
  - name: etcd
    image: mysql
  - name: init1
    image: quay.io/cybozu/init
`

		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(testCtx, po)
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})

	It("should deny a Pod having untrustworthy Ephemeral Container", func() {
		podManifest := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod-validate-4
  namespace: default
spec:
  containers:
  - name: ubuntu
    image: quay.io/cybozu/ubuntu
`

		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(testCtx, po)
		Expect(err).NotTo(HaveOccurred())

		ephemeralContainer := v1.EphemeralContainer{
			EphemeralContainerCommon: v1.EphemeralContainerCommon{
				Name:  "debug",
				Image: "busybox",
			},
		}
		po.Spec.EphemeralContainers = append(po.Spec.EphemeralContainers, ephemeralContainer)

		k8s, err := kubernetes.NewForConfig(k8sConfig)
		Expect(err).NotTo(HaveOccurred())

		po, err = k8s.CoreV1().Pods(po.Namespace).UpdateEphemeralContainers(testCtx, po.Name, po, metav1.UpdateOptions{})
		permissive := os.Getenv("TEST_PERMISSIVE") == "true"
		Expect(err == nil).To(Equal(permissive))
	})
})
