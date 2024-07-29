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
	// **************************************************リファクタリングすること
	f, err := os.Open(path.Join(ptrDir, serial))
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(path.Join(ptrDir, serial))
		if err != nil {
			slog.Error(fmt.Sprintf("%s", err))
			return lptr, err
		}
		lptr := LastPointer{
			Serial:       serial,
			LastReadTime: 0,
			LastReadId:   0,
		}
		f.Close()
		return lptr, err
	} else if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return lptr, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return lptr, err
	}
	if st.Size() == 0 {
		return lptr, nil
	}
	byteJSON, err := io.ReadAll(f)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return lptr, err
	}
	if json.Unmarshal(byteJSON, &lptr) != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return lptr, err
	}
	//mutex.Unlock()

	return lptr, err
}

func (c *logCollector) updateLastPointer(lptr LastPointer, ptrDir string) error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	file, err := os.Create(path.Join(ptrDir, lptr.Serial))
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	_, err = file.WriteString(string(byteJSON))
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	//mutex.Unlock()

	return nil
}
