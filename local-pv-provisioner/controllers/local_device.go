package controllers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func (dd *DeviceDetector) listLocalDevices(ctx context.Context) ([]device, error) {
	log := dd.log.WithValues("node", dd.nodeName)
	var devs []device

	err := filepath.Walk(dd.deviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err, "prevent panic by handling failure accessing", "path", path)
			return err
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		if dd.deviceNameFilter.MatchString(filepath.Base(path)) {
			capacityBytes, err := dd.getCapacityBytes(ctx, path)
			if err != nil {
				log.Error(err, "unable to get capacity", "path", path)
				return err
			}
			devs = append(devs, device{name: path, capacityBytes: capacityBytes})
		}
		return nil
	})
	if err != nil {
		log.Error(err, "error while walking the path", "path", dd.deviceDir)
		return nil, err
	}
	return devs, nil
}

func (dd *DeviceDetector) getCapacityBytes(ctx context.Context, path string) (int64, error) {
	out, err := exec.CommandContext(ctx, "lsblk", path, "--bytes", "--output=SIZE", "--noheadings").Output()
	if err != nil {
		return 0, err
	}
	capacity, err := strconv.ParseInt(string(out), 10, 64)
	if err != nil {
		return 0, err
	}
	return capacity, nil
}
