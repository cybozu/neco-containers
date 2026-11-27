package main

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

func TestUpdate(t *testing.T) {
	output := filepath.Join(t.TempDir(), "resolve-map.json")
	if err := update([]Machine{newMachine("CS1", "192.0.2.4")}, output); err != nil {
		t.Fatal(err)
	}

	actual, err := resolvemap.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"CS1": "192.0.2.4"}
	if !maps.Equal(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

// TestUpdateOnlyWhenChanged checks that the file is written only when the content is changed.
func TestUpdateOnlyWhenChanged(t *testing.T) {
	output := filepath.Join(t.TempDir(), "resolve-map.json")
	if err := update([]Machine{newMachine("CS1", "192.0.2.4")}, output); err != nil {
		t.Fatal(err)
	}

	// Set an old modification time to detect whether the file is rewritten.
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(output, old, old); err != nil {
		t.Fatal(err)
	}
	modTime := func() time.Time {
		t.Helper()
		fi, err := os.Stat(output)
		if err != nil {
			t.Fatal(err)
		}
		return fi.ModTime()
	}

	// not changed
	if err := update([]Machine{newMachine("CS1", "192.0.2.4")}, output); err != nil {
		t.Fatal(err)
	}
	if !modTime().Equal(old) {
		t.Error("the file should not be rewritten if the content is not changed")
	}

	// changed
	if err := update([]Machine{newMachine("CS1", "192.0.2.5")}, output); err != nil {
		t.Fatal(err)
	}
	if modTime().Equal(old) {
		t.Error("the file should be rewritten if the content is changed")
	}
	actual, err := resolvemap.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	if actual["CS1"] != "192.0.2.5" {
		t.Errorf("the file is not updated: %v", actual)
	}
}

func TestUpdateKeepsFileOnError(t *testing.T) {
	cases := []struct {
		name     string
		machines []Machine
	}{
		// sabakan returns no machines, e.g. 404 from the REST API.
		{name: "NoMachines", machines: nil},
		{name: "NoBMCAddresses", machines: []Machine{newMachine("CS1", "")}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// The existing file must be kept on errors.
			output := filepath.Join(t.TempDir(), "resolve-map.json")
			old := map[string]string{"foo": "203.0.113.1"}
			if err := resolvemap.Write(output, old); err != nil {
				t.Fatal(err)
			}
			if err := update(c.machines, output); err == nil {
				t.Fatal("expected an error")
			}
			actual, err := resolvemap.Read(output)
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(actual, old) {
				t.Errorf("the file was changed: %v", actual)
			}
		})
	}
}

func newMachine(serial, bmcIPv4 string) Machine {
	var m Machine
	m.Spec.Serial = serial
	m.Spec.BMC.IPv4 = bmcIPv4
	return m
}

func TestResolveMapData(t *testing.T) {
	machines := []Machine{
		newMachine("ABC1234", "192.0.2.3"),
		newMachine("CS1", "192.0.2.4"),
		// sabakan returns serials in upper case. Check the conversion anyway.
		newMachine("ss2", "192.0.2.5"),
		// skipped: no BMC address
		newMachine("NOBMC", ""),
		// skipped: no serial
		newMachine("", "192.0.2.6"),
	}

	expected := map[string]string{
		"ABC1234": "192.0.2.3",
		"CS1":     "192.0.2.4",
		"SS2":     "192.0.2.5",
	}

	actual := resolveMapData(machines)
	if !maps.Equal(actual, expected) {
		t.Errorf("unexpected resolve map:\nexpected: %v\nactual:   %v", expected, actual)
	}
}
