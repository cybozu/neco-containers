package main

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// EndpointSlice
	defaultMaxEndpointsPerSlice = 100

	// target names
	targetAllServersName  = "all-servers-targets"
	targetBootServersName = "boot-servers-targets"
)

// validateMaxEndpointsPerSlice checks that maxEndpointsPerSlice is in the
// range accepted by EndpointSlice.Endpoints.
func validateMaxEndpointsPerSlice(maxEndpointsPerSlice int) error {
	if maxEndpointsPerSlice <= 0 || maxEndpointsPerSlice > 1000 {
		return fmt.Errorf("max-endpoints-per-slice %d is out of range", maxEndpointsPerSlice)
	}
	return nil
}

// parseNamedPorts parses specs of the form "name:port" into namedPorts,
// validating that name is a valid Service/EndpointPort port name, that port
// is in the range 1-65535, and that names are unique. The result is sorted
// by name, so that it is deterministic regardless of input order.
func parseNamedPorts(specs []string) ([]namedPort, error) {
	seen := make(map[string]struct{}, len(specs))
	ports := make([]namedPort, 0, len(specs))
	for _, spec := range specs {
		name, portStr, ok := strings.Cut(spec, ":")
		if !ok {
			return nil, fmt.Errorf("invalid port %q (expected name:port)", spec)
		}
		if errs := validation.IsValidPortName(name); len(errs) > 0 {
			return nil, fmt.Errorf("invalid port name %q: %s", name, strings.Join(errs, "; "))
		}
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", spec, err)
		}
		if port == 0 {
			return nil, fmt.Errorf("invalid port %q: port must be between 1 and 65535", spec)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate port name %q", name)
		}
		seen[name] = struct{}{}
		ports = append(ports, namedPort{name: name, port: int32(port)})
	}
	slices.SortFunc(ports, func(a, b namedPort) int { return strings.Compare(a.name, b.name) })
	return ports, nil
}

// allServerIPs returns the IPs of all non-retired machines with an IPv4 address.
// Machines with an unparsable IPv4 address are skipped with a warning log,
// so that a single bad sabakan record doesn't block unrelated processing.
func allServerIPs(machines []Machine) []net.IP {
	var ips []net.IP
	for _, machine := range machines {
		if machine.isRetired() || !machine.hasIPv4() {
			continue
		}
		ip, ok := parseMachineIPv4(machine)
		if !ok {
			continue
		}
		ips = append(ips, ip)
	}
	return ips
}

// bootServerIPs returns the IPs of non-retired boot servers with an IPv4 address.
// Machines with an unparsable IPv4 address are skipped with a warning log,
// so that a single bad sabakan record doesn't block unrelated processing.
func bootServerIPs(machines []Machine) []net.IP {
	var ips []net.IP
	for _, machine := range machines {
		if machine.isRetired() || !machine.hasIPv4() || !machine.isBootServer() {
			continue
		}
		ip, ok := parseMachineIPv4(machine)
		if !ok {
			continue
		}
		ips = append(ips, ip)
	}
	return ips
}

func parseMachineIPv4(machine Machine) (net.IP, bool) {
	ip := net.ParseIP(machine.Spec.IPv4[0])
	if ip == nil {
		slog.Warn("skipping machine with an invalid IPv4 address", "serial", machine.Spec.Serial, "ipv4", machine.Spec.IPv4[0])
		return nil, false
	}
	return ip, true
}
