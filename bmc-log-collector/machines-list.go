package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Machine struct {
	Serial string
	BmcIP  string
	NodeIP string
}

type Machines struct {
	machine []Machine
}

// Get iDRAC server list from CSV file
func machineListReader(filename string) (Machines, error) {
	var mlist Machines
	file, err := os.Open(filename)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return mlist, err
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	for {
		items, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error(fmt.Sprintf("%s", err))
			return mlist, err
		}
		n := len(items)
		if n != 3 {
			err := fmt.Errorf("invalid machine list CSV, number of items = %d", len(items))
			slog.Error(fmt.Sprintf("%s", err))
			return mlist, err
		}
		mlist.machine = append(mlist.machine, Machine{Serial: items[0], BmcIP: items[1], NodeIP: items[2]})
	}
	return mlist, nil
}
