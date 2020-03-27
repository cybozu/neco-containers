package updateblock117

import "testing"

func TestExistsDeviceFile(t *testing.T) {
	exists, err := existsDeviceFile("/dev/null")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("cannot find /dev/null")
	}

	exists, err = existsDeviceFile("/etc/hosts")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("/etc/hosts is unexpectedly detected as a device file")
	}
}
