package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"time"
)

type LastPointer struct {
	Serial       string
	LastReadTime int64
	LastReadId   int
	//LastUpdateTime int64
	LastError error
}

func readLastPointer(serial string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer

	filePath := path.Join(ptrDir, serial)
	f, err := os.Open(filePath)
	// when new file, create file and set initial data.
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(filePath)
		if err != nil {
			slog.Error("os.Create()", "err", err, "filename", filePath)
			return lptr, err
		}
		lptr := LastPointer{
			Serial:       serial,
			LastReadTime: 0,
			LastReadId:   0,
			//LastUpdateTime: time.Now().Unix(),
		}
		f.Close()
		return lptr, err
		// when other error occur
	} else if err != nil {
		slog.Error("os.Open()", "err", err, "filename", filePath)
		return lptr, err
	}
	defer f.Close()

	byteJSON, err := io.ReadAll(f)
	if err != nil {
		slog.Error("io.ReadAll()", "err", err, "filename", filePath)
		return lptr, err
	}

	if json.Unmarshal(byteJSON, &lptr) != nil {
		slog.Error("json.Unmarshal()", "err", err, "byteJSON", string(byteJSON))
		return lptr, err
	}
	return lptr, err
}

func updateLastPointer(lptr LastPointer, ptrDir string) error {
	filePath := path.Join(ptrDir, lptr.Serial)
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("os.Create()", "err", err, "filename", filePath)
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		slog.Error("json.Marshal()", "err", err)
		return err
	}
	_, err = file.WriteString(string(byteJSON))
	if err != nil {
		slog.Error("file.WriteString()", "err", err, "writing data", string(byteJSON))
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
