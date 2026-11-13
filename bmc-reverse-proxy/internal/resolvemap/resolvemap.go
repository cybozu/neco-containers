// Package resolvemap defines the resolve map file shared by update-resolve-map and bmc-reverse-proxy.
//
// The file is a JSON object that maps names to BMC addresses, e.g.
//
//	{"rack1-cs4": "192.0.2.4", "198-51-100-196": "192.0.2.4", "ABC1234": "192.0.2.4"}
//
// update-resolve-map writes it, and bmc-reverse-proxy reads it.
package resolvemap

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DefaultPath is the default path of the resolve map file.
const DefaultPath = "/var/lib/resolve-map/resolve-map.json"

// Read reads the resolve map from the file at path.
// The error from os.ReadFile is returned as is, so that the caller can check fs.ErrNotExist.
// A file containing null is an error, so that a nil map always means "not loaded" for the caller.
func Read(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("resolve map is null")
	}
	return m, nil
}

// Write writes the resolve map to the file at path.
// The file is replaced atomically, so that readers never see a partially written file.
func Write(path string, m map[string]string) (err error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Create the temporary file in the same directory so that os.Rename does not cross file systems.
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()           //nolint:errcheck
			os.Remove(f.Name()) //nolint:errcheck
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Chmod(0o644); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
