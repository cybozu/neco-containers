package sabakan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MachineState string

type MachineBMC struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Type string `json:"type"`
}

// MachineSpec is a set of attributes to define a machine.
type MachineSpec struct {
	Serial       string            `json:"serial"`
	Labels       map[string]string `json:"labels"`
	Rack         uint              `json:"rack"`
	IndexInRack  uint              `json:"index-in-rack"`
	Role         string            `json:"role"`
	IPv4         []string          `json:"ipv4"`
	IPv6         []string          `json:"ipv6"`
	RegisterDate time.Time         `json:"register-date"`
	RetireDate   time.Time         `json:"retire-date"`
	BMC          MachineBMC        `json:"bmc"`
}

// MachineStatus represents the status of a machine.
type MachineStatus struct {
	Timestamp time.Time    `json:"timestamp"`
	Duration  float64      `json:"duration"`
	State     MachineState `json:"state"`
}

// NetworkInfo represents NIC configurations.
type NetworkInfo struct {
	IPv4 []NICConfig `json:"ipv4"`
}

// BMCInfo represents BMC NIC configuration information.
type BMCInfo struct {
	IPv4 NICConfig `json:"ipv4"`
}

// NICConfig represents NIC configuration information.
type NICConfig struct {
	Address  string `json:"address"`
	Netmask  string `json:"netmask"`
	MaskBits int    `json:"maskbits"`
	Gateway  string `json:"gateway"`
}

// MachineInfo is a set of associated information of a Machine.
type MachineInfo struct {
	Network NetworkInfo `json:"network"`
	BMC     BMCInfo     `json:"bmc"`
}

// Machine represents a server hardware.
type Machine struct {
	Spec   MachineSpec   `json:"spec"`
	Status MachineStatus `json:"status"`
	Info   MachineInfo   `json:"info"`
}

func GetBmcIpv4(sabakanEndpoint string, serial string) (string, error) {
	req, _ := http.NewRequest("GET", sabakanEndpoint+"/"+serial, nil)
	client := &http.Client{Timeout: time.Duration(3) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// NotFound の時は？
	byteJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(byteJSON) == 0 {
		return "", fmt.Errorf("serial is not found in neco/sabakan")
	}

	var machines []Machine
	err = json.Unmarshal(byteJSON, &machines)
	if err != nil {
		return "", nil
	}
	return machines[0].Info.BMC.IPv4.Address, nil
}
