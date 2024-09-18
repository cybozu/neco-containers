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
	stringJSON, err := createMachinesList(ml)
	if err != nil {
		t.Fatalf("failed create JSON data %#v", err)
	}
	if !cmp.Equal(stringJSON, expectedJSON) {
		t.Fatalf("Not expected JSON data %v", expectedJSON)
	}
}
