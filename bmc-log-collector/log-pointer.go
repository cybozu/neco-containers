package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
)

type LastPointer struct {
	LastReadId         int    // BMC SEL Id
	LastError          string // for TCP Error
	LastHttpStatusCode int    // for HTTP Error
	FirstCreateTime    int64

	// Lifecycle log (files written by older versions lack these fields,
	// so they are loaded as zero values and the LC log starts fresh)
	LcLastReadId         int    // BMC LC log Id
	LcLastReadCreateTime int64  // Created time of the LcLastReadId entry, to detect log clear
	LcLastError          string // for TCP Error
	LcLastHttpStatusCode int    // for HTTP Error
}

func checkAndCreatePointerFile(filePath string) error {
	var lptr LastPointer

	// If the pointer file does not exist, create a new one.
	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		return nil
	}

	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		return err
	}

	fd, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.WriteString(string(byteJSON))
	if err != nil {
		return err
	}
	return err
}

func updateLastPointer(lptr LastPointer, filePath string) error {
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	dirName := filepath.Dir(filePath)
	tmpPath := dirName + "/_" + fileName
	fd, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.WriteString(string(byteJSON))
	if err != nil {
		return err
	}

	err = fd.Sync()
	if err != nil {
		return err
	}

	err = os.Rename(tmpPath, filePath)
	if err != nil {
		return err
	}
	return nil
}

// loadLastPointer reads the pointer file of a machine, creating an empty one
// when the machine is seen for the first time.
func loadLastPointer(filePath string) (LastPointer, error) {
	if err := checkAndCreatePointerFile(filePath); err != nil {
		return LastPointer{}, fmt.Errorf("check pointer file: %w", err)
	}
	lastPtr, err := readLastPointer(filePath)
	if err != nil {
		return LastPointer{}, fmt.Errorf("read pointer file: %w", err)
	}
	return lastPtr, nil
}

func readLastPointer(filePath string) (LastPointer, error) {
	var lptr LastPointer
	fd, err := os.Open(filePath)
	if err != nil {
		return lptr, err
	}
	defer fd.Close()
	byteJSON, err := io.ReadAll(fd)
	if err != nil {
		return lptr, err
	}
	err = json.Unmarshal(byteJSON, &lptr)
	if err != nil {
		return lptr, err
	}
	return lptr, err
}

// Delete pointer files when disappear at machines list which from ConfigMap
func deletePtrFileDisappearedSerial(ptrDir string, machinesList []Machine) error {
	machineExist := make(map[string]bool, len(machinesList))
	for _, m := range machinesList {
		machineExist[m.Serial] = true
	}

	files, err := os.ReadDir(ptrDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Remove the pointer file and metrics
		if !machineExist[file.Name()] {
			// Remove the pointer file
			filePath := path.Join(ptrDir, file.Name())
			err = os.Remove(filePath)
			if err != nil {
				slog.Error("failed to remove the pointer file", "err", err, "file", filePath)
			}
			// Clear metrics counter
			deleteMetrics(file.Name())
		}
	}
	return nil
}
