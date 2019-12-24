package controllers

import (
	"io"
	"os"
	"path/filepath"
)

func (dd *DeviceDetector) listLocalDevices() ([]Device, error) {
	log := dd.log.WithValues("node", dd.nodeName)
	var devs []Device

	err := filepath.Walk(dd.deviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err, "prevent panic by handling failure accessing", "path", path)
			return err
		}
		log.Info("walk path", "path", path)

		if info.IsDir() {
			return nil
		}

		if dd.deviceNameFilter.MatchString(filepath.Base(path)) {
			capacityBytes, err := dd.getCapacityBytes(path)
			if err != nil {
				log.Error(err, "unable to get capacity", "path", path)
				return err
			}
			log.Info("get capacity", "path", path, "capacity", capacityBytes)
			devs = append(devs, Device{Path: path, CapacityBytes: capacityBytes})
		}
		return nil
	})
	if err != nil {
		log.Error(err, "error while walking the path", "path", dd.deviceDir)
		return nil, err
	}
	return devs, nil
}

func (dd *DeviceDetector) getCapacityBytes(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	return pos, nil
}
