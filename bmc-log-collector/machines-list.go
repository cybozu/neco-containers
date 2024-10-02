package main

import (
	"encoding/json"
	"io"
	"os"
)

type Machine struct {
	Serial string `json:"serial"`
	BmcIP  string `json:"bmc_ipv4"`
	NodeIP string `json:"node_ipv4"`
}

// Get BMC list from JSON file
func readMachineList(filename string) ([]Machine, error) {
	var ml []Machine

	fd, err := os.Open(filename)
	if err != nil {
		return ml, err
	}
	defer fd.Close()

	byteData, err := io.ReadAll(fd)
	if err != nil {
		return ml, err
	}

	err = json.Unmarshal(byteData, &ml)
	if err != nil {
		return ml, err
	}

	return ml, nil
}
