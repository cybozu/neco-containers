package hooks

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
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
	It("should append volumeMount to container", func() {
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
		Expect(out.Spec.Volumes).Should(HaveLen(1))
		Expect(out.Spec.Volumes[0].VolumeSource).Should(Equal(v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}))
		Expect(out.Spec.Containers[0].VolumeMounts).Should(HaveLen(1))
		Expect(out.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal("/tmp"))
	})

	It("should append volumeMount to initContainer", func() {
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
		Expect(out.Spec.Volumes).Should(HaveLen(2))
		Expect(out.Spec.Volumes[0].VolumeSource).Should(Equal(v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}))
		Expect(out.Spec.Volumes[1].VolumeSource).Should(Equal(v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}))
		Expect(out.Spec.Containers[0].VolumeMounts).Should(HaveLen(1))
		Expect(out.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal("/tmp"))
		Expect(out.Spec.InitContainers[0].VolumeMounts).Should(HaveLen(1))
		Expect(out.Spec.InitContainers[0].VolumeMounts[0].MountPath).Should(Equal("/tmp"))
	})

	It("should not append volumeMount to containers", func() {
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
   volumeMounts:
   - name: vol1
     mountPath: /tmp/hoge
 initContainers:
 - name: init
   image: quay.io/cybozu/ubuntu
   command: ["pause"]
   volumeMounts:
   - name: vol2
     mountPath: /tmp
 volumes:
 - name: vol1
   emptyDir: {}
 - name: vol2
   emptyDir: {}
`
		d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(podManifest), 4096)
		po := &v1.Pod{}
		err := d.Decode(po)
		Expect(err).NotTo(HaveOccurred())

		out := createPod(po)
		Expect(out.Spec.Volumes).Should(HaveLen(2))
		Expect(out.Spec.Volumes[0].VolumeSource).Should(Equal(v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}))
		Expect(out.Spec.Volumes[1].VolumeSource).Should(Equal(v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}))
		Expect(out.Spec.Containers[0].VolumeMounts).Should(HaveLen(1))
		Expect(out.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal("/tmp/hoge"))
		Expect(out.Spec.InitContainers[0].VolumeMounts).Should(HaveLen(1))
		Expect(out.Spec.InitContainers[0].VolumeMounts[0].MountPath).Should(Equal("/tmp"))
	})
})

func TestPodMutatorIsMountedTmp(t *testing.T) {
	tests := []struct {
		name      string
		container *v1.Container
		want      bool
	}{
		{
			name:      "empty volumeMounts",
			container: &v1.Container{VolumeMounts: []v1.VolumeMount{}},
			want:      false,
		},
		{
			name:      "/tmp",
			container: &v1.Container{VolumeMounts: []v1.VolumeMount{{MountPath: "/tmp"}}},
			want:      true,
		},
		{
			name:      "/tmp/hoge",
			container: &v1.Container{VolumeMounts: []v1.VolumeMount{{MountPath: "/tmp/hoge"}}},
			want:      true,
		},
		{
			name:      "/tmp1",
			container: &v1.Container{VolumeMounts: []v1.VolumeMount{{MountPath: "/tmp1"}}},
			want:      false,
		},
		{
			name:      "/hoge, /piyo, /tmp",
			container: &v1.Container{VolumeMounts: []v1.VolumeMount{{MountPath: "/hoge"}, {MountPath: "/piyo"}, {MountPath: "/tmp"}}},
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &podMutator{}
			if got := m.isMountedTmp(tt.container); got != tt.want {
				t.Errorf("podMutator.isMountedTmp() = %v, want %v", got, tt.want)
			}
		})
	}
}
