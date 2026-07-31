package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaselineCacheRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sub", "baseline")

	if err := SaveBaselineCache(path, mustParse(t, "0-1,6-11")); err != nil {
		t.Fatal(err)
	}
	got, err := LoadBaselineCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "0-1,6-11"; got.String() != want {
		t.Errorf("LoadBaselineCache = %q, want %q", got.String(), want)
	}

	// Overwrite with a new value.
	if err := SaveBaselineCache(path, mustParse(t, "0-3")); err != nil {
		t.Fatal(err)
	}
	got, err = LoadBaselineCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "0-3"; got.String() != want {
		t.Errorf("LoadBaselineCache after overwrite = %q, want %q", got.String(), want)
	}
}

func TestLoadBaselineCacheErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		if _, err := LoadBaselineCache(filepath.Join(t.TempDir(), "nope")); err == nil {
			t.Error("LoadBaselineCache succeeded, want error")
		}
	})

	t.Run("garbage content", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "baseline")
		if err := os.WriteFile(path, []byte("not-a-cpuset"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadBaselineCache(path); err == nil {
			t.Error("LoadBaselineCache succeeded, want error")
		}
	})
}
