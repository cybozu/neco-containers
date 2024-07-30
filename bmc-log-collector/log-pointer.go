package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

type LastPointer struct {
	Serial       string
	LastReadTime int64
	LastReadId   int
}

func (c *logCollector) readLastPointer(serial string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer

	c.mutex.Lock()
	defer c.mutex.Unlock()

	filePath := path.Join(ptrDir, serial)
	f, err := os.Open(filePath)
	// when new file, create file and set initial data.
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(filePath)
		if err != nil {
			slog.Error(fmt.Sprintf("%s, %s", err, filePath))
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
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return lptr, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return lptr, err
	}

	if st.Size() == 0 {
		return lptr, nil
	}

	byteJSON, err := io.ReadAll(f)
	if err != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return lptr, err
	}

	if json.Unmarshal(byteJSON, &lptr) != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return lptr, err
	}

	return lptr, err
}

func (c *logCollector) updateLastPointer(lptr LastPointer, ptrDir string) error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	filePath := path.Join(ptrDir, lptr.Serial)
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return err
	}
	_, err = file.WriteString(string(byteJSON))
	if err != nil {
		slog.Error(fmt.Sprintf("%s, %s", err, filePath))
		return err
	}
	return nil
}
