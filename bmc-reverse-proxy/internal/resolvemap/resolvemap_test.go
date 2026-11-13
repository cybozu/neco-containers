package resolvemap

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	m, err := Read(write("valid.json", `{"rack1-cs4": "192.0.2.4", "CS1": "192.0.2.4"}`))
	if err != nil {
		t.Fatal(err)
	}
	if expected := map[string]string{"rack1-cs4": "192.0.2.4", "CS1": "192.0.2.4"}; !maps.Equal(m, expected) {
		t.Errorf("expected %v, got %v", expected, m)
	}

	m, err = Read(write("empty.json", `{}`))
	if err != nil || m == nil || len(m) != 0 {
		t.Errorf("an empty object should be read as an empty map: %v, %v", m, err)
	}

	if _, err := Read(filepath.Join(dir, "not-exist.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected fs.ErrNotExist, got %v", err)
	}
	if _, err := Read(write("broken.json", `{`)); err == nil {
		t.Error("expected an error for broken JSON")
	}
	if _, err := Read(write("null.json", `null`)); err == nil {
		t.Error("expected an error for null")
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "resolve-map.json")

	for _, m := range []map[string]string{
		{"rack1-cs4": "192.0.2.4", "CS1": "192.0.2.4"},
		// overwrite an existing file
		{"boot-0": "192.0.2.3"},
	} {
		if err := Write(path, m); err != nil {
			t.Fatal(err)
		}
		actual, err := Read(path)
		if err != nil {
			t.Fatal(err)
		}
		if !maps.Equal(actual, m) {
			t.Errorf("expected %v, got %v", m, actual)
		}
	}

	// no temporary file should be left
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("unexpected files: %v", entries)
	}

	if err := Write(filepath.Join(dir, "no-such-dir", "resolve-map.json"), map[string]string{}); err == nil {
		t.Error("expected error for a missing directory")
	}
}
