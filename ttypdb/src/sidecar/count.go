package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var errProcStat = errors.New("broken process stat")

// ttyCount returns the number of controlling terminals observed.
// NOTE: This implementation is for Linux.
func ttyCount() (int, error) {
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}

	ttys := map[string]bool{}
	for _, d := range dirs {
		err := func() error {
			name := d.Name()
			for _, ch := range name {
				if ch < '0' || ch > '9' {
					return nil
				}
			}

			statBytes, err := os.ReadFile(filepath.Join("/proc", name, "stat"))
			if err != nil {
				return err
			}

			// The 6th (0-origin) field is controlling tty device number.
			// If it is "0", the process is not controlled.
			stat := string(statBytes)
			idx := strings.LastIndexByte(stat, ')')
			if idx == -1 {
				return errProcStat
			}
			stat = stat[idx+2:] // two fields are skipped
			fields := strings.SplitN(stat, " ", 6)
			if len(fields) <= 4 {
				return errProcStat
			}
			ttyNumber := fields[4]
			if ttyNumber != "0" {
				ttys[ttyNumber] = true
			}
			return nil
		}()
		if err != nil {
			return 0, err
		}
	}

	return len(ttys), nil
}
