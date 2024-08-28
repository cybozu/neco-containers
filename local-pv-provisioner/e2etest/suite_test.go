package e2etest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kubectlPath = os.Getenv("KUBECTL")
)

func execAtLocal(cmd string, input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if len(input) != 0 {
		command.Stdin = bytes.NewReader(input)
	}

	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func kubectl(args ...string) ([]byte, []byte, error) {
	if len(kubectlPath) == 0 {
		panic("KUBECTL environment variable is not set")
	}
	return execAtLocal(kubectlPath, nil, args...)
}

func getObject[T any](kind, namespace, name string) (*T, error) {
	var stdout []byte
	var err error

	if namespace == "" {
		stdout, _, err = kubectl("get", kind, name, "-o", "json")
	} else {
		stdout, _, err = kubectl("get", kind, "-n", namespace, name, "-o", "json")
	}
	if err != nil {
		return nil, err
	}

	var obj T
	if err := json.Unmarshal(stdout, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func getPV(name string) (*corev1.PersistentVolume, error) {
	return getObject[corev1.PersistentVolume]("persistentvolume", "", name)
}

func TestMtest(t *testing.T) {
	if os.Getenv("E2ETEST") == "" {
		t.Skip("Run under e2e/")
	}

	RegisterFailHandler(Fail)

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(3 * time.Minute)

	RunSpecs(t, "local-pv-provisioner tests")
}

var _ = Describe("local-pv-provisioner", func() {
	Context("smoke tests", func() {
		Describe("make sure that PVs can be created and deleted correctly", func() {
			AfterEach(func() {
				By("cleaning up")
				_, _, _ = kubectl("annotate", "node", "minikube-worker", "local-pv-provisioner.cybozu.io/pv-spec-configmap-")
				_, _, _ = kubectl("delete", "pv", "local-minikube-worker-loop0")
				_, _, _ = kubectl("delete", "pv", "local-minikube-worker-loop1")
			})

			DescribeTable(
				"checking that PVs should be created and deleted correctly",
				func(pvSpecConfigMap string, expectedVolumeMode corev1.PersistentVolumeMode, testPodManifestPath string) {
					By("adding annotation to Node")
					_, _, err := kubectl("annotate", "node", "minikube-worker",
						"local-pv-provisioner.cybozu.io/pv-spec-configmap="+pvSpecConfigMap)
					Expect(err).NotTo(HaveOccurred())

					By("checking that PVs are created")
					Eventually(func() error {
						pv0, err := getPV("local-minikube-worker-loop0")
						if err != nil {
							return err
						}
						pv1, err := getPV("local-minikube-worker-loop1")
						if err != nil {
							return err
						}
						Expect(pv0.Status.Phase).To(Equal(corev1.VolumeAvailable))
						Expect(*pv0.Spec.VolumeMode).To(Equal(expectedVolumeMode))
						Expect(pv1.Status.Phase).To(Equal(corev1.VolumeAvailable))
						Expect(*pv1.Spec.VolumeMode).To(Equal(expectedVolumeMode))
						return nil
					}).Should(Succeed())

					By("starting a Pod using the PV")
					_, _, err = kubectl("apply", "-f", testPodManifestPath)
					Expect(err).NotTo(HaveOccurred())

					By("checking that one PV becomes Bound")
					var pvBound corev1.PersistentVolume
					Eventually(func() error {
						pv0, err := getPV("local-minikube-worker-loop0")
						if err != nil {
							return err
						}
						if pv0.Status.Phase == corev1.VolumeBound {
							pvBound = *pv0
							return nil
						}
						pv1, err := getPV("local-minikube-worker-loop1")
						if err != nil {
							return err
						}
						if pv1.Status.Phase == corev1.VolumeBound {
							pvBound = *pv1
							return nil
						}
						return errors.New("not bound yet")
					}).Should(Succeed())

					By("deleting the Pod")
					_, _, err = kubectl("delete", "-f", testPodManifestPath)
					Expect(err).NotTo(HaveOccurred())

					By("checking that the Bound PV becomes Available")
					Eventually(func() error {
						pv, err := getPV(pvBound.GetName())
						if err != nil {
							return err
						}
						if pv.Status.Phase == corev1.VolumeAvailable {
							return nil
						}
						return errors.New("not available yet")
					}).Should(Succeed())
				},
				Entry("Block", "pv-spec-cm-block", corev1.PersistentVolumeBlock, "testdata/test-pod-block.yaml"),
				Entry("Filesystem", "pv-spec-cm-fs", corev1.PersistentVolumeFilesystem, "testdata/test-pod-fs.yaml"),
			)
		})
	})
})