package controllers

import (
	"io"
	"path/filepath"
	"regexp"
)

func (dd *DeviceDetector) listLocalDevices(deviceDir string, deviceNameFilter *regexp.Regexp) ([]Device, []Device, error) {
	log := dd.log
	var devs []Device
	var errDevs []Device

	err := fs.Walk(deviceDir, func(path string, info FileInfo, err error) error {
		if err != nil {
			log.Error(err, "failure accessing a path", "path", path)
			return err
		}

		if info.IsDir() {
			return nil
		}

		if deviceNameFilter.MatchString(filepath.Base(path)) {
			capacityBytes, err := dd.getCapacityBytes(path)
			if err != nil {
				log.Error(err, "unable to get capacity", "path", path)
				errDevs = append(errDevs, Device{Path: path})
			} else {
				devs = append(devs, Device{Path: path, CapacityBytes: capacityBytes})
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err, "error while walking the path", "path", deviceDir)
		return nil, nil, err
	}

	return devs, errDevs, nil
}

func (dd *DeviceDetector) getCapacityBytes(path string) (int64, error) {
	file, err := fs.Open(path)
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
