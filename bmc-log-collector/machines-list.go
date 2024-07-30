package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Machine struct {
	Serial string `json:"serial"`
	BmcIP  string `json:"bmc_ip"`
	NodeIP string `json:"ipv4"`
}

type Machines struct {
	Machine []Machine
}

// get BMC list from JSON file
func machineListReader(filename string) (Machines, error) {
	var ml Machines

	file, err := os.Open(filename)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return ml, err
	}
	defer file.Close()

	byteData, err := io.ReadAll(file)
	if err != nil {
		return ml, err
	}

	err = json.Unmarshal(byteData, &ml)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return ml, err
	}

	return ml, nil
}
