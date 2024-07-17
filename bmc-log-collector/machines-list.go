package main

import (
	"encoding/csv"
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
		slog.Error("failed open file")
		return mlist, err
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	for {
		item, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed read file")
			return mlist, err
		}
		mlist.machine = append(mlist.machine, Machine{Serial: item[0], BmcIP: item[1], NodeIP: item[2]})
	}
	return mlist, nil
}
