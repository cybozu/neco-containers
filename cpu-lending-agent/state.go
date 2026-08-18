package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// CPUManagerState is the part of the kubelet CPU manager checkpoint
// (/var/lib/kubelet/cpu_manager_state) that the agent depends on.
//
// The agent treats this file as the read-only source of truth for exclusive
// CPU assignments and the shared pool. The checksum is intentionally ignored;
// the kubelet is the only writer.
type CPUManagerState struct {
	PolicyName    string                       `json:"policyName"`
	DefaultCPUSet string                       `json:"defaultCpuSet"`
	Entries       map[string]map[string]string `json:"entries"`
}

// PinnedState is the parsed form of CPUManagerState.
type PinnedState struct {
	Static     bool
	SharedPool CPUSet
	// Reserved maps pod UID -> container name -> exclusive CPUs.
	Reserved map[string]map[string]CPUSet
}

// LoadPinnedState reads and parses the kubelet CPU manager checkpoint.
func LoadPinnedState(path string) (*PinnedState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var raw CPUManagerState
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	shared, err := ParseCPUSet(raw.DefaultCPUSet)
	if err != nil {
		return nil, fmt.Errorf("invalid defaultCpuSet in %s: %w", path, err)
	}
	st := &PinnedState{
		Static:     raw.PolicyName == "static",
		SharedPool: shared,
		Reserved:   map[string]map[string]CPUSet{},
	}
	for podUID, containers := range raw.Entries {
		for name, cpus := range containers {
			set, err := ParseCPUSet(cpus)
			if err != nil {
				return nil, fmt.Errorf("invalid cpuset for pod %s container %s in %s: %w", podUID, name, path, err)
			}
			if st.Reserved[podUID] == nil {
				st.Reserved[podUID] = map[string]CPUSet{}
			}
			st.Reserved[podUID][name] = set
		}
	}
	return st, nil
}
