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
	LastReadTime int64
	LastReadId   int
	OffSet       int
}

// 排他制御を入れること！！
func readLastPointer(serial string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer
	f, err := os.Open(path.Join(ptrDir, serial))
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(path.Join(ptrDir, serial))
		if err != nil {
			//slog.Error("failed to create pointer file")
			return lptr, err
		}
		lptr := LastPointer{
			Serial:       serial,
			LastReadTime: 0,
			LastReadId:   0,
			OffSet:       0,
		}
		f.Close()
		return lptr, err
	} else if err != nil {
		//slog.Error("failed to open pointer file")
		return lptr, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		//slog.Error("failed to get the status of the file")
		return lptr, err
	}
	if st.Size() == 0 {
		return lptr, nil
	}
	byteJSON, err := io.ReadAll(f)
	if err != nil {
		//slog.Error("failed to read pointer file")
		return lptr, err
	}
	if json.Unmarshal(byteJSON, &lptr) != nil {
		//slog.Error("failed to convert the struct from JSON")
		return lptr, err
	}
	return lptr, err
}

func updateLastPointer(lptr LastPointer, ptrDir string) error {
	file, err := os.Create(path.Join(ptrDir, lptr.Serial))
	if err != nil {
		//slog.Error("failed to open pointer file")
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		//slog.Error("failed to convert JSON")
		return err
	}
	n, err := file.WriteString(string(byteJSON))
	if err != nil { //|| n == 0 {
		//slog.Error("failed to save the log pointer")
		fmt.Println("wrote bytes=", n)
		return err
	}
	return nil
}
