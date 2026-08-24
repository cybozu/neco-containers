package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseNamedPorts(t *testing.T) {
	ports, err := parseNamedPorts([]string{"node-exporter:9100", "etcd-metrics:2381"})
	if err != nil {
		t.Fatal(err)
	}
	// sorted by name, regardless of input order
	want := []namedPort{
		{name: "etcd-metrics", port: 2381},
		{name: "node-exporter", port: 9100},
	}
	if !cmp.Equal(ports, want, cmp.AllowUnexported(namedPort{})) {
		t.Errorf("unexpected ports: %#v", ports)
	}

	if _, err := parseNamedPorts([]string{"invalid"}); err == nil {
		t.Error("expected an error for a spec without a port")
	}
	if _, err := parseNamedPorts([]string{"name:not-a-number"}); err == nil {
		t.Error("expected an error for a non-numeric port")
	}
	if _, err := parseNamedPorts([]string{"Invalid-Name:1234"}); err == nil {
		t.Error("expected an error for an invalid port name")
	}
	if _, err := parseNamedPorts([]string{"a-name-too-long-to-be-valid:1234"}); err == nil {
		t.Error("expected an error for a port name longer than 15 characters")
	}
	if _, err := parseNamedPorts([]string{"12345:1234"}); err == nil {
		t.Error("expected an error for a port name without any letters")
	}
	if _, err := parseNamedPorts([]string{"name:0"}); err == nil {
		t.Error("expected an error for port 0")
	}
	if _, err := parseNamedPorts([]string{"name:65536"}); err == nil {
		t.Error("expected an error for a port out of range")
	}
	if _, err := parseNamedPorts([]string{"name:1234", "name:5678"}); err == nil {
		t.Error("expected an error for a duplicate port name")
	}
}

func TestValidateMaxEndpointsPerSlice(t *testing.T) {
	if err := validateMaxEndpointsPerSlice(100); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := validateMaxEndpointsPerSlice(0); err == nil {
		t.Error("expected an error for 0")
	}
	if err := validateMaxEndpointsPerSlice(1001); err == nil {
		t.Error("expected an error for a value out of range")
	}
}

func TestServerIPs(t *testing.T) {
	var boot, worker, retired Machine
	boot.Spec.IPv4 = []string{"1.1.1.1"}
	boot.Spec.Role = "boot"
	worker.Spec.IPv4 = []string{"2.2.2.2"}
	worker.Spec.Role = "worker"
	retired.Spec.IPv4 = []string{"3.3.3.3"}
	retired.Status.State = "RETIRED"
	machines := []Machine{boot, worker, retired}

	all := allServerIPs(machines)
	if len(all) != 2 {
		t.Errorf("unexpected allServerIPs: %v", all)
	}

	boots := bootServerIPs(machines)
	if len(boots) != 1 || boots[0].String() != "1.1.1.1" {
		t.Errorf("unexpected bootServerIPs: %v", boots)
	}

	var invalid Machine
	invalid.Spec.IPv4 = []string{"not-an-ip"}
	invalid.Spec.Role = "boot"
	if ips := allServerIPs([]Machine{invalid}); len(ips) != 0 {
		t.Errorf("expected an invalid IPv4 address to be skipped, got: %v", ips)
	}
	if ips := bootServerIPs([]Machine{invalid}); len(ips) != 0 {
		t.Errorf("expected an invalid IPv4 address to be skipped, got: %v", ips)
	}
}
