package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUpdateBMCLogCollectorConfigMap(t *testing.T) {
	var ml []Machine

	var m0 Machine
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.1.1.1")
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.2.2.2")
	m0.Spec.BMC.IPv4 = "1.3.3.3"
	m0.Spec.Serial = "ABC123"
	ml = append(ml, m0)

	var m1 Machine
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.1.1.1")
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.2.2.2")
	m1.Spec.BMC.IPv4 = "2.3.3.3"
	m1.Spec.Serial = "XYZ123"
	ml = append(ml, m1)

	// expectedJSON is made from ml
	expectedJSON := `[{"serial":"ABC123","bmc_ipv4":"1.3.3.3","node_ipv4":"1.1.1.1"},{"serial":"XYZ123","bmc_ipv4":"2.3.3.3","node_ipv4":"2.1.1.1"}]`
	data, err := bmcLogCollectorConfigMapData(ml)
	if err != nil {
		t.Fatalf("failed create JSON data %#v", err)
	}
	if !cmp.Equal(data["machineslist.json"], expectedJSON) {
		t.Fatalf("Not expected JSON data %v", expectedJSON)
	}
}

func TestBMCConfigMapDataSkipsMachinesWithoutNodeIPv4(t *testing.T) {
	var noNodeIPv4 Machine
	noNodeIPv4.Spec.BMC.IPv4 = "1.3.3.3"
	noNodeIPv4.Spec.Serial = "ABC123"
	machines := []Machine{noNodeIPv4}

	addresses := bmcReverseProxyConfigMapData(machines)
	if len(addresses) != 0 {
		t.Errorf("expected no addresses for a machine without a node IPv4, got %#v", addresses)
	}

	data, err := bmcLogCollectorConfigMapData(machines)
	if err != nil {
		t.Fatal(err)
	}
	if data["machineslist.json"] != "null" {
		t.Errorf("expected an empty list for a machine without a node IPv4, got %q", data["machineslist.json"])
	}
}
