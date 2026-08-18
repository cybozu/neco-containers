package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeState(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cpu_manager_state")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadPinnedState(t *testing.T) {
	t.Parallel()

	t.Run("static with entries", func(t *testing.T) {
		t.Parallel()
		// Taken from a real kubelet v1.34 checkpoint.
		path := writeState(t, `{"policyName":"static","defaultCpuSet":"0-1,6-11","entries":{"4aec46af-22b7-4e26-babb-d15383ac9151":{"mysqld":"2-3"},"60008d73-6f6f-4467-96c5-cd0a1eeadf23":{"mysqld2":"4-5"}},"checksum":2764424633}`)
		st, err := LoadPinnedState(path)
		if err != nil {
			t.Fatal(err)
		}
		if !st.Static {
			t.Error("Static = false, want true")
		}
		if got, want := st.SharedPool.String(), "0-1,6-11"; got != want {
			t.Errorf("SharedPool = %q, want %q", got, want)
		}
		if got, want := st.Reserved["4aec46af-22b7-4e26-babb-d15383ac9151"]["mysqld"].String(), "2-3"; got != want {
			t.Errorf("Reserved = %q, want %q", got, want)
		}
	})

	t.Run("multiple containers in one pod are kept separate", func(t *testing.T) {
		t.Parallel()
		path := writeState(t, `{"policyName":"static","defaultCpuSet":"0-1","entries":{"uid1":{"c1":"2-3","c2":"4-5"}}}`)
		st, err := LoadPinnedState(path)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := st.Reserved["uid1"]["c1"].String(), "2-3"; got != want {
			t.Errorf("Reserved[c1] = %q, want %q", got, want)
		}
		if got, want := st.Reserved["uid1"]["c2"].String(), "4-5"; got != want {
			t.Errorf("Reserved[c2] = %q, want %q", got, want)
		}
	})

	t.Run("non-static policy", func(t *testing.T) {
		t.Parallel()
		path := writeState(t, `{"policyName":"none","defaultCpuSet":"","checksum":1353318690}`)
		st, err := LoadPinnedState(path)
		if err != nil {
			t.Fatal(err)
		}
		if st.Static {
			t.Error("Static = true, want false")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		if _, err := LoadPinnedState(filepath.Join(t.TempDir(), "nope")); err == nil {
			t.Error("LoadPinnedState succeeded, want error")
		}
	})

	t.Run("broken json", func(t *testing.T) {
		t.Parallel()
		path := writeState(t, `{"policyName":`)
		if _, err := LoadPinnedState(path); err == nil {
			t.Error("LoadPinnedState succeeded, want error")
		}
	})
}
