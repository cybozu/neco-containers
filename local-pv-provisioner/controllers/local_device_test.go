package controllers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestDeviceDetectorListLocalDevices(t *testing.T) {
	tests := []struct {
		name             string
		deviceNameFilter string
		inputDevices     []Device
		want             []Device
		wantErrDev       []Device
		wantErr          bool
	}{
		{
			name:             "no device",
			deviceNameFilter: ".*",
			inputDevices:     []Device{},
			want:             []Device{},
			wantErrDev:       []Device{},
			wantErr:          false,
		},
		{
			name:             "exist a device",
			deviceNameFilter: ".*",
			inputDevices:     []Device{{Path: "dev01", CapacityBytes: 512}},
			want:             []Device{{Path: "dev01", CapacityBytes: 512}},
			wantErrDev:       []Device{},
			wantErr:          false,
		},
		{
			name:             "exist multiple devices",
			deviceNameFilter: ".*",
			inputDevices:     []Device{{Path: "dev01", CapacityBytes: 512}, {Path: "dev02", CapacityBytes: 1024}},
			want:             []Device{{Path: "dev01", CapacityBytes: 512}, {Path: "dev02", CapacityBytes: 1024}},
			wantErrDev:       []Device{},
			wantErr:          false,
		},
		{
			name:             "exist error devices",
			deviceNameFilter: ".*",
			inputDevices: []Device{
				{Path: "dev01", CapacityBytes: 512},
				{Path: "dev02", CapacityBytes: 1024},
				{Path: "dev03", CapacityBytes: 0},
				{Path: "dev04", CapacityBytes: 0},
			},
			want:       []Device{{Path: "dev01", CapacityBytes: 512}, {Path: "dev02", CapacityBytes: 1024}},
			wantErrDev: []Device{{Path: "dev03", CapacityBytes: 0}, {Path: "dev04", CapacityBytes: 0}},
			wantErr:    false,
		},
		{
			name:             "filter specified devices",
			deviceNameFilter: "^dev",
			inputDevices:     []Device{{Path: "foo", CapacityBytes: 512}, {Path: "dev01", CapacityBytes: 512}},
			want:             []Device{{Path: "dev01", CapacityBytes: 512}},
			wantErrDev:       []Device{},
			wantErr:          false,
		},
	}

	log := zap.New(zap.UseDevMode(true))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyFileDir, symlinkDir := setupDummyDevice(t, tt.inputDevices)
			defer os.RemoveAll(dummyFileDir)
			defer os.RemoveAll(symlinkDir)

			dd := &DeviceDetector{
				log:              log,
				deviceDir:        symlinkDir,
				deviceNameFilter: regexp.MustCompile(tt.deviceNameFilter),
				nodeName:         "test-node-name",
			}
			got, errDev, err := dd.listLocalDevices()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeviceDetector.listLocalDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len(got) != len(tt.want): %d != %d", len(got), len(tt.want))
			}
			for i, dev := range got {
				if filepath.Base(dev.Path) != tt.want[i].Path {
					t.Errorf("filepath.Base(dev.path) != tt.want[i].path: %s != %s", filepath.Base(dev.Path), tt.want[i].Path)
				}
				if dev.CapacityBytes != tt.want[i].CapacityBytes {
					t.Errorf("dev.capacityBytes != tt.want[i].capacityBytes: %d != %d", dev.CapacityBytes, tt.want[i].CapacityBytes)
				}
			}
			if len(errDev) != len(tt.wantErrDev) {
				t.Fatalf("len(errDev) != len(tt.wantErrDev): %d != %d", len(errDev), len(tt.wantErrDev))
			}
			for i, dev := range errDev {
				if filepath.Base(dev.Path) != tt.wantErrDev[i].Path {
					t.Errorf("filepath.Base(dev.path) != tt.wantErrDev[i].path: %s != %s", filepath.Base(dev.Path), tt.wantErrDev[i].Path)
				}
				if dev.CapacityBytes != tt.wantErrDev[i].CapacityBytes {
					t.Errorf("dev.capacityBytes != tt.wantErrDev[i].capacityBytes: %d != %d", dev.CapacityBytes, tt.wantErrDev[i].CapacityBytes)
				}
			}
		})
	}
}

func setupDummyDevice(t *testing.T, devices []Device) (string, string) {
	dummyFileDir, err := ioutil.TempDir("", "list-local-devices-dummy-")
	if err != nil {
		t.Fatal(err)
	}

	symlinkDir, err := ioutil.TempDir("", "list-local-devices-symlink-")
	if err != nil {
		t.Fatal(err)
	}

	for _, device := range devices {
		dummyDeviceName := filepath.Join(dummyFileDir, device.Path+".dummy")
		dummyDeviceSymlink := filepath.Join(symlinkDir, device.Path)

		if device.CapacityBytes != 0 {
			err := exec.Command("fallocate", "-l", fmt.Sprintf("%d", device.CapacityBytes), dummyDeviceName).Run()
			if err != nil {
				t.Fatal(err)
			}
		}

		err = exec.Command("ln", "-s", dummyDeviceName, dummyDeviceSymlink).Run()
		if err != nil {
			t.Fatal(err)
		}
	}
	return dummyFileDir, symlinkDir
}
