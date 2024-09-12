package main

import (
	"encoding/json"
	"testing"
)

func TestUpdateBMCLogCollectorConfigMap(t *testing.T) {

	var ml []Machine
	var m0 Machine
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.1.1.1")
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.2.2.2")
	m0.Spec.BMC.IPv4 = "1.3.3.3"
	m0.Spec.Serial = "ABC123"
	m0.Spec.Role = "cs"
	m0.Status.State = "HEALTHY"
	ml = append(ml, m0)

	var m1 Machine
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.1.1.1")
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.2.2.2")
	m1.Spec.BMC.IPv4 = "2.3.3.3"
	m1.Spec.Serial = "XYZ123"
	m1.Spec.Role = "boot"
	m1.Status.State = "HEALTHY"
	ml = append(ml, m1)

	byteJSON, err := bmcListJson(ml)
	if err != nil {
		t.Fatalf("failed create JSON data %#v", err)
	}

	type MachineConfigMap struct {
		Serial string `json:"serial"`
		BmcIP  string `json:"bmc_ip"`
		NodeIP string `json:"ipv4"`
		Role   string `json:"role"`
		State  string `json:"state"`
	}

	var mcm []MachineConfigMap
	byteArray := []byte(byteJSON["serverlist.json"])
	if err := json.Unmarshal(byteArray, &mcm); err != nil {
		t.Fatalf("failed convert JSON data %#v", err)
	}

	{
		if mcm[0].Serial != m0.Spec.Serial {
			t.Fatalf("failed convert Serial %v", mcm[0].Serial)
		}
		if mcm[0].BmcIP != m0.Spec.BMC.IPv4 {
			t.Fatalf("failed convert BmcIP %v", mcm[0].BmcIP)
		}
		if mcm[0].NodeIP != m0.Spec.IPv4[0] {
			t.Fatalf("failed convert NodeIP %v", mcm[0].NodeIP)
		}
		if mcm[0].Role != m0.Spec.Role {
			t.Fatalf("failed convert Role %v", mcm[0].Role)
		}
		if mcm[0].State != m0.Status.State {
			t.Fatalf("failed convert State %v", mcm[0].State)
		}
	}

	{
		if mcm[1].Serial != m1.Spec.Serial {
			t.Fatalf("failed convert Serial %v", mcm[1].Serial)
		}
		if mcm[1].BmcIP != m1.Spec.BMC.IPv4 {
			t.Fatalf("failed convert BmcIP %v", mcm[1].BmcIP)
		}
		if mcm[1].NodeIP != m1.Spec.IPv4[0] {
			t.Fatalf("failed convert NodeIP %v", mcm[1].NodeIP)
		}
		if mcm[1].Role != m1.Spec.Role {
			t.Fatalf("failed convert Role %v", mcm[1].Role)
		}
		if mcm[1].State != m1.Status.State {
			t.Fatalf("failed convert State %v", mcm[1].State)
		}
	}

}
