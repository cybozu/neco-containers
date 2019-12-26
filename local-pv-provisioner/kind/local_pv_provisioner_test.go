package kind_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cybozu/neco-containers/local-pv-provisioner/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	countNode         = 2
	deviceSymlinkPath = "/dev/crypt-disk/by-path/"
)

var cleanupTargetDevice []string

func setupDevice(devices []controllers.Device) ([]string, error) {
	td, err := ioutil.TempDir("", "list-local-devices-dummy-")
	if err != nil {
		return nil, fmt.Errorf("failed to TempDir %w", err)
	}

	err = exec.Command("sudo", "mkdir", "-p", deviceSymlinkPath).Run()
	if err != nil {
		return nil, fmt.Errorf("failed to mkdir %w", err)
	}

	var loopDevices []string
	for _, device := range devices {
		dummyFileName := filepath.Join(td, filepath.Base(device.Path)+".dummy")

		err := exec.Command("fallocate", "-l", fmt.Sprintf("%d", device.CapacityBytes), dummyFileName).Run()
		if err != nil {
			return nil, fmt.Errorf("failed to fallocate %w", err)
		}

		out, err := exec.Command("sudo", "losetup", "-f", dummyFileName, "--show").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to losetup %w", err)
		}

		loopDevicePath := strings.TrimSpace(string(out))
		err = exec.Command("sudo", "ln", "-s", loopDevicePath, device.Path).Run()
		if err != nil {
			return nil, fmt.Errorf("failed to ln %s %s %w", loopDevicePath, device.Path, err)
		}

		err = exec.Command("sudo", "chmod", "755", device.Path).Run()
		if err != nil {
			return nil, fmt.Errorf("failed to chmod %w", err)
		}

		loopDevices = append(loopDevices, loopDevicePath)
	}
	return loopDevices, nil
}

var _ = Describe("test local-pv-provisioner", func() {
	It("should create PV from device in worker nodes", func() {
		By("setup devices")
		devices := []controllers.Device{
			{
				Path:          filepath.Join(deviceSymlinkPath, "crypt-dev-01"),
				CapacityBytes: 512,
			},
			{
				Path:          filepath.Join(deviceSymlinkPath, "crypt-dev-02"),
				CapacityBytes: 512,
			},
		}
		loopDevices, err := setupDevice(devices)
		Expect(err).ShouldNot(HaveOccurred())
		cleanupTargetDevice = append(cleanupTargetDevice, loopDevices...)

		By("deploying provisioner daemonset")
		err = exec.Command("kubectl", "apply", "-f", "install.yaml").Run()
		Expect(err).ShouldNot(HaveOccurred())

		By("confirming that PVs are created")
		Eventually(func() error {
			out, err := exec.Command("kubectl", "get", "pv", "-o", "json").Output()
			if err != nil {
				return err
			}

			var pvl corev1.PersistentVolumeList
			err = json.Unmarshal(out, &pvl)
			if err != nil {
				return err
			}
			if len(pvl.Items) != len(devices)*countNode {
				return errors.New("len(pvl.Items) != len(devices) * countNode")
			}
			for _, item := range pvl.Items {
				quantity, err := resource.ParseQuantity("512")
				if err != nil {
					return err
				}
				if item.Spec.Capacity["storage"] != quantity {
					return errors.New(`item.Spec.Capacity["storage"] != quantity`)
				}
			}
			return nil
		}).Should(Succeed())

		By("deleting worker node")
		Expect(exec.Command("kubectl", "delete", "node", "lpp-test-worker").Run()).ShouldNot(HaveOccurred())
		Eventually(func() error {
			out, err := exec.Command("kubectl", "get", "pv", "-o", "json").Output()
			if err != nil {
				return err
			}

			var pvl corev1.PersistentVolumeList
			err = json.Unmarshal(out, &pvl)
			if err != nil {
				return err
			}
			if len(pvl.Items) != len(devices)*(countNode-1) {
				return errors.New("len(pvl.Items) != len(devices) * (countNode-1)")
			}
			return nil
		}).Should(Succeed())
	})
})
