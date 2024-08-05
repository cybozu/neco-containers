package main

import (
	"encoding/json"
	"io"
	"os"
)

type Machine struct {
	Serial string `json:"serial"`
	BmcIP  string `json:"bmc_ip"`
	NodeIP string `json:"ipv4"`
}

// get BMC list from JSON file
func machineListReader(filename string) ([]Machine, error) {
	var ml []Machine

	file, err := os.Open(filename)
	if err != nil {
		return ml, err
	}
	defer file.Close()

	byteData, err := io.ReadAll(file)
	if err != nil {
		return ml, err
	}

	err = json.Unmarshal(byteData, &ml)
	if err != nil {
		return ml, err
	}

	return ml, nil
}
