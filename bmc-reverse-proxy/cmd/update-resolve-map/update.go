package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

// update gets the machines from sabakan and writes the resolve map file.
func update(ctx context.Context, sabakan sabakanClient, output string) error {
	machines, err := sabakan.getMachines(ctx)
	if err != nil {
		return err
	}
	data := resolveMapData(machines)
	if len(data) == 0 {
		// sabakan should always have machines. Keep the current file rather than
		// making all BMCs unreachable because of a wrong response from sabakan.
		return fmt.Errorf("no BMC addresses found in %d machines from sabakan; keep the current resolve map", len(machines))
	}
	if err := resolvemap.Write(output, data); err != nil {
		return fmt.Errorf("failed to write resolve map: %w", err)
	}
	slog.Info("updated resolve map", "output", output, "machines", len(machines), "entries", len(data))
	return nil
}

// resolveMapData returns the map from hostnames to BMC addresses.
// The keys are exactly the names that clients can use as the first label of the host,
// like "rack1-cs4" in "rack1-cs4.bmc.<cluster>.<base>".
// Hostnames are in lower case as in DNS, and serials are in upper case as shown by sabakan and BMCs.
// bmc-reverse-proxy looks up the map case-insensitively.
//
// Each machine is registered with three keys: its hostname, its node address with dots replaced by hyphens,
// and its serial. Retired machines are also included because they are operated via BMCs.
func resolveMapData(machines []Machine) map[string]string {
	addresses := make(map[string]string)
	for _, machine := range machines {
		if machine.Spec.BMC.IPv4 == "" || !machine.hasIPv4() {
			continue
		}

		var hostname string
		if machine.isBootServer() {
			// Though the full hostname of a boot server is like "stage0-boot-0",
			// the cluster name "stage0" is not available here. So use "boot-0".
			hostname = fmt.Sprintf("boot-%d", machine.Spec.Rack)
		} else {
			hostname = fmt.Sprintf("rack%d-%s%d", machine.Spec.Rack, machine.Spec.Role, machine.Spec.IndexInRack)
		}

		// "a.b.c.d" does not match the wildcard in "*.bmc.<cluster>.<base>".  "a-b-c-d" does match.
		nodeAddress := strings.ReplaceAll(machine.Spec.IPv4[0], ".", "-")

		addresses[strings.ToLower(hostname)] = machine.Spec.BMC.IPv4
		addresses[strings.ToLower(nodeAddress)] = machine.Spec.BMC.IPv4
		addresses[strings.ToUpper(machine.Spec.Serial)] = machine.Spec.BMC.IPv4
	}
	return addresses
}
