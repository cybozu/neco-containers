package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveBaselineCache atomically persists the shared pool to path.
func SaveBaselineCache(path string, baseline CPUSet) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create baseline cache directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create baseline cache temp file: %w", err)
	}
	defer func() {
		// Best-effort cleanup; after a successful rename the file is gone.
		_ = os.Remove(tmp.Name())
	}()
	if _, err := tmp.WriteString(baseline.String() + "\n"); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write baseline cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close baseline cache temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("failed to commit baseline cache: %w", err)
	}
	return nil
}

// LoadBaselineCache reads the shared pool persisted by SaveBaselineCache.
func LoadBaselineCache(path string) (CPUSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline cache: %w", err)
	}
	set, err := ParseCPUSet(string(data))
	if err != nil {
		return nil, fmt.Errorf("invalid baseline cache %s: %w", path, err)
	}
	return set, nil
}
