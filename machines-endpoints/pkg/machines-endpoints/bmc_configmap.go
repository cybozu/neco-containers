package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// BMC Reverse Proxy ConfigMap
	bmcReverseProxyConfigMapName = "bmc-reverse-proxy"

	// BMC Log Collector ConfigMap
	bmcLogCollectorConfigMapName = "bmc-log-collector"
)

func bmcReverseProxyConfigMapData(machines []Machine) map[string]string {
	addresses := make(map[string]string)
	for _, machine := range machines {
		if machine.Spec.BMC.IPv4 == "" || !machine.hasIPv4() {
			continue
		}

		var hostname string
		if machine.isBootServer() {
			// Though full hostname is like "stage0-boot-0",
			// the part of "stage0-" is insignificant in a cluster while it is hard to get.
			// So use "boot-0" for resolving.
			hostname = fmt.Sprintf("boot-%d", machine.Spec.Rack)
		} else {
			hostname = fmt.Sprintf("rack%d-%s%d", machine.Spec.Rack, machine.Spec.Role, machine.Spec.IndexInRack)
		}
		addresses[hostname] = machine.Spec.BMC.IPv4

		// "a.b.c.d" does not match the wildcard in "*.bmc.<cluster>.<base>".  "a-b-c-d" does match.
		addresses[strings.ReplaceAll(machine.Spec.IPv4[0], ".", "-")] = machine.Spec.BMC.IPv4

		addresses[machine.Spec.Serial] = machine.Spec.BMC.IPv4
	}

	return addresses
}

func bmcLogCollectorConfigMapData(machines []Machine) (map[string]string, error) {
	type machineSerialAndIP struct {
		Serial string `json:"serial"`
		BmcIP  string `json:"bmc_ipv4"`
		NodeIP string `json:"node_ipv4"`
	}

	var ml []machineSerialAndIP
	for _, machine := range machines {
		if machine.Spec.BMC.IPv4 == "" || !machine.hasIPv4() {
			continue
		}
		ml = append(ml, machineSerialAndIP{
			Serial: machine.Spec.Serial,
			BmcIP:  machine.Spec.BMC.IPv4,
			NodeIP: machine.Spec.IPv4[0],
		})
	}

	byteJSON, err := json.Marshal(ml)
	if err != nil {
		return nil, err
	}
	return map[string]string{"machineslist.json": string(byteJSON)}, nil
}
