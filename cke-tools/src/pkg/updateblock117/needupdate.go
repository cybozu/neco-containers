package updateblock117

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	deviceRootDir    = "/var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/"
	devicePublishDir = "/var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/publish/"
)

func makeOldDeviceFilePath(pvName string) string {
	return filepath.Join(devicePublishDir, pvName)
}

func makeTmpDeviceFilePath(pvName string) string {
	return filepath.Join(devicePublishDir, pvName+".tmp")
}

func makeNewDeviceFilePath(pvName, podUID string) string {
	return filepath.Join(devicePublishDir, pvName, podUID)
}

// ExistsBlockDeviceAtOldLocation returns true if the PV's path is located at old path
// old path: `plugins/kubernetes.io/csi/volumeDevices/publish/{pvname}`
// new path: `plugins/kubernetes.io/csi/volumeDevices/publish/{pvname}/{podUid}`
// ref https://github.com/kubernetes/kubernetes/pull/74026
func ExistsBlockDeviceAtOldLocation(pvName string) (bool, error) {
	oldPath := makeOldDeviceFilePath(pvName)
	return existsDeviceFile(oldPath)
}

// ExistsBlockDeviceAtTmp returns true if the PV's path is located at /tmp/{pvname}.
func ExistsBlockDeviceAtTmp(pvName string) (bool, error) {
	tmpPath := makeTmpDeviceFilePath(pvName)
	return existsDeviceFile(tmpPath)
}

func existsDeviceFile(location string) (bool, error) {
	fi, err := os.Stat(location)
	if err != nil {
		return false, err
	}

	isDevice := fi.Mode()&os.ModeDevice != 0
	return isDevice, nil
}

// IsSymlinkOutdated returns true if the PV's symlink does not point at
// new path `plugins/kubernetes.io/csi/volumeDevices/publish/{pvname}/{podUid}`.
func IsSymlinkOutdated(pvName string) (bool, error) {
	podUID, err := getPodUID(pvName)
	if err != nil {
		return false, err
	}

	symlinkPath := filepath.Join(deviceRootDir, pvName, "dev", podUID)
	res, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, err
	}
	outdated := res != makeNewDeviceFilePath(pvName, podUID)
	return outdated, nil
}

// getPodUid returns pod UID bound with the PV.
// The symlink is located at `plugins/kubernetes.io/csi/volumeDevices/{pvname}/dev/{podUid}`.
func getPodUID(pvName string) (string, error) {
	symlinkDir := filepath.Join(deviceRootDir, pvName, "dev")
	var podUID string
	files, err := ioutil.ReadDir(symlinkDir)
	if err != nil {
		return "", err
	}
	for _, info := range files {
		isSymlink := info.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			continue
		}
		podUID = info.Name()
		break
	}
	if len(podUID) == 0 {
		return "", errors.New("symlink of " + pvName + " is not found")
	}
	return podUID, nil
}
