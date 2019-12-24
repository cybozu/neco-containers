package kind_test

import (
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(10 * time.Second)

	RunSpecs(t, "Kind Suite")
}

var _ = AfterSuite(func() {
	err := exec.Command("kubectl", "delete", "-f", "install.yaml").Run()
	if err != nil {
		log.Log.Error(err, "failed to kubectl delete -f install.yaml")
	}

	err = exec.Command("kubectl", "delete", "pv", "--all").Run()
	if err != nil {
		log.Log.Error(err, "failed to kubectl delete pv")
	}

	err = cleanupDevice(cleanupTargetDevice, deviceSymlinkPath)
	if err != nil {
		log.Log.Error(err, "cleanupDevice")
	}
})

func cleanupDevice(loopDevices []string, deviceDir string) error {
	err := exec.Command("sudo", "rm", "-r", deviceDir).Run()
	if err != nil {
		return err
	}
	for _, loopDevice := range loopDevices {
		err := exec.Command("sudo", "losetup", "-d", loopDevice).Run()
		if err != nil {
			return err
		}
	}
	return nil
}
