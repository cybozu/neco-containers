package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type LastPointer struct {
	Serial       string
	NodeIP       string
	LastReadTime int64
	LastReadId   int
	LastError    error
}

func readLastPointer(serial string, nodeIp string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer

	filePath := path.Join(ptrDir, serial)
	f, err := os.Open(filePath)
	// when new file, create file and set initial data.
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(filePath)
		if err != nil {
			return lptr, err
		}
		lptr := LastPointer{
			Serial:       serial,
			NodeIP:       nodeIp,
			LastReadTime: 0,
			LastReadId:   0,
		}
		f.Close()
		return lptr, err
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
	return lptr, err
}

func updateLastPointer(lptr LastPointer, ptrDir string) error {
	filePath := path.Join(ptrDir, lptr.Serial)
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

// Delete pointer files that have not been updated
func deleteUnUpdatedFiles(ptrDir string, machinesList []Machine) error {

	machines := make(map[string]Machine, len(machinesList))
	for _, m := range machinesList {
		machines[m.Serial] = m
	}

	files, err := os.ReadDir(ptrDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fmt.Println("================", file.Name())

		_, isExist := machines[file.Name()]
		fmt.Println("check", isExist)
		if !isExist {
			filePath := path.Join(ptrDir, file.Name())
			os.Remove(filePath)
		}
	}

	return nil
}

/*
func getMachineListWhichEverAccessed(ptrDir string) (map[string]LastPointer, error) {
	files, err := os.ReadDir(ptrDir)
	if err != nil {
		return nil, err
	}
	machineList := make(map[string]LastPointer, len(files))

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := path.Join(ptrDir, file.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		f, err := os.Open(file.Name())
		if err != nil {
			return nil, err
		}
		defer f.Close()

		byteJSON, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		var machine LastPointer
		if json.Unmarshal(byteJSON, &machine) != nil {
			return nil, err
		}
		machineList[machine.Serial] = machine
	}

	return machineList, nil
}
*/
