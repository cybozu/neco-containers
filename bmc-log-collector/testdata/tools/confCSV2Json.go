package tool

import (
	"encoding/csv"
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

// Get iDRAC server list from CSV file
func machineListReader(filename string) (Machines, error) {
	var m Machines
	file, err := os.Open(filename)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return m, err
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
			return m, err
		}
		n := len(items)
		if n != 3 {
			err := fmt.Errorf("invalid machine list CSV, number of items = %d", len(items))
			slog.Error(fmt.Sprintf("%s", err))
			return m, err
		}
		m.Machine = append(m.Machine, Machine{Serial: items[0], BmcIP: items[1], NodeIP: items[2]})
	}
	return m, nil
}

func convJSONtext(m Machines, fn string) error {
	file, err := os.Create(fn)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(m)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	_, err = file.WriteString(string(byteJSON))
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return err
	}
	return nil
}

func main() {
	var m Machines
	m, err := machineListReader(os.Args[1])
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return
	}
	convJSONtext(m, os.Args[2])
}
