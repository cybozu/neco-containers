package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
)

type LastPointer struct {
	LastReadTime int64
	LastReadId   int
	LastError    error
}

// func readLastPointer(serial string, nodeIp string, ptrDir string) (LastPointer, error) {
func readLastPointer(serial string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer
	filePath := path.Join(ptrDir, serial)
	f, err := os.Open(filePath)
	// If the file does not exist, create file and set initial data.
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(filePath)
		if err != nil {
			return lptr, err
		}
		lptr := LastPointer{
			LastReadTime: 0,
			LastReadId:   0,
		}
		f.Close()
		return lptr, nil
		// when other error occur
	} else if err != nil {
		return lptr, err
	}
	defer f.Close()

	byteJSON, err := io.ReadAll(f)
	if err != nil {
		return lptr, err
	}

	if json.Unmarshal(byteJSON, &lptr) != nil {
		return lptr, err
	}
	return lptr, nil
}

func updateLastPointer(lptr LastPointer, ptrDir string, serial string) error {
	filePath := path.Join(ptrDir, serial)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(byteJSON))
	if err != nil {
		return err
	}
	return nil
}

// Delete pointer files when disappear machines list which from ConfigMap
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
		if !machineExist[file.Name()] {
			filePath := path.Join(ptrDir, file.Name())
			os.Remove(filePath)
		}
	}
	return nil
}
