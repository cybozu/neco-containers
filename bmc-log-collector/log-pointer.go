package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"time"
)

type LastPointer struct {
	Serial       string
	LastReadTime int64
	LastReadId   int
	LastError    error
}

func readLastPointer(serial string, ptrDir string) (LastPointer, error) {
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
func deleteUnUpdatedFiles(ptrDir string) error {

	files, err := os.ReadDir(ptrDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := path.Join(ptrDir, file.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		st, err := file.Stat()
		if err != nil {
			return err
		}

		// Remove a file that no update for 6 months
		if (time.Now().Unix() - st.ModTime().Unix()) >= 3600*24*30*6 {
			os.Remove(file.Name())
		}
	}
	return nil
}
