package updateblock117

import (
	"os"
	"path/filepath"
)

// MoveBlockDeviceToTmp moves block device to tmp path.
func MoveBlockDeviceToTmp(pvName string) error {
	oldPath := filepath.Join(devicePublishDir, pvName)
	tmpPath := filepath.Join("/tmp", pvName)
	return os.Rename(oldPath, tmpPath)
}

// MoveBlockDeviceToNew moves block device to new path.
func MoveBlockDeviceToNew(pvName string) error {
	podUID, err := getPodUID(pvName)
	if err != nil {
		return err
	}

	dir := filepath.Join(devicePublishDir, pvName)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}

	tmpPath := filepath.Join("/tmp", pvName)
	newPath := filepath.Join(dir, podUID)
	return os.Rename(tmpPath, newPath)
}

// UpdateSymlink updates symlink destination.
func UpdateSymlink(pvName string) error {
	podUID, err := getPodUID(pvName)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(devicePublishDir, pvName, podUID)
	symlinkPath := filepath.Join(deviceRootDir, pvName, "dev", podUID)
	symlinkTmpPath := symlinkPath + ".tmp"

	err = os.Remove(symlinkTmpPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Symlink(targetPath, symlinkTmpPath)
	if err != nil {
		return err
	}

	return os.Rename(symlinkTmpPath, symlinkPath)
}
