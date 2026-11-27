package main

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

// update writes the resolve map file built from the machines.
// It does nothing if the content is not changed. So it does not log every update.
func update(machines []Machine, output string) error {
	data := resolveMapData(machines)
	if len(data) == 0 {
		// sabakan should always have machines. A wrong response must not make all BMCs unreachable.
		// So keep the current file.
		return fmt.Errorf("no BMC addresses found in %d machines from sabakan; keep the current resolve map", len(machines))
	}
	// Write the file anyway if the current one cannot be read, e.g. it does not exist yet.
	if current, err := resolvemap.Read(output); err == nil && maps.Equal(current, data) {
		return nil
	}
	if err := resolvemap.Write(output, data); err != nil {
		return fmt.Errorf("failed to write resolve map: %w", err)
	}
	slog.Info("updated resolve map", "output", output, "machines", len(machines), "entries", len(data))
	return nil
}

// resolveMapData returns the map from serials to BMC addresses.
// Clients use a serial as the first label of the host, e.g. "abc1234.bmc.example.com".
// The serials are in upper case, as shown by sabakan and BMCs.
// bmc-reverse-proxy looks up the map case-insensitively.
// Retired machines are also included, because they are operated via BMCs.
func resolveMapData(machines []Machine) map[string]string {
	addresses := make(map[string]string)
	for _, machine := range machines {
		if machine.Spec.Serial == "" || machine.Spec.BMC.IPv4 == "" {
			continue
		}
		addresses[strings.ToUpper(machine.Spec.Serial)] = machine.Spec.BMC.IPv4
	}
	return addresses
}
