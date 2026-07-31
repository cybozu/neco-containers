package main

import (
	"fmt"
	"os"
	"sync"
)

// Topology resolves SMT siblings from sysfs. Safe for concurrent use
// (SiblingsOf is reachable from both the NRI Synchronize path and the
// reconcile loop).
type Topology struct {
	// SysRoot is normally "/sys"; overridable for tests.
	SysRoot string

	mu       sync.Mutex
	siblings map[int]CPUSet
}

// NewTopology creates a Topology reading from sysRoot.
func NewTopology(sysRoot string) *Topology {
	return &Topology{SysRoot: sysRoot, siblings: map[int]CPUSet{}}
}

// SiblingsOf returns the SMT sibling set of cpu (including cpu itself).
// Results are cached; the CPU topology does not change at runtime.
func (t *Topology) SiblingsOf(cpu int) (CPUSet, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.siblings[cpu]; ok {
		return s, nil
	}
	path := fmt.Sprintf("%s/devices/system/cpu/cpu%d/topology/thread_siblings_list", t.SysRoot, cpu)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SMT topology: %w", err)
	}
	s, err := ParseCPUSet(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SMT topology %s: %w", path, err)
	}
	t.siblings[cpu] = s
	return s, nil
}

// WholeCores filters set down to CPUs whose SMT siblings are all contained
// in allowed. In lending terms: a CPU may be lent only when the borrower
// could also run on every sibling (because it is lent too, or already part
// of the shared pool). This keeps borrowers off physical cores whose other
// thread belongs to someone else's exclusive assignment, without dropping
// CPUs whose sibling the borrower can use anyway.
func (t *Topology) WholeCores(set, allowed CPUSet) (CPUSet, error) {
	out := CPUSet{}
	for cpu := range set {
		sib, err := t.SiblingsOf(cpu)
		if err != nil {
			return nil, err
		}
		whole := true
		for s := range sib {
			if !allowed.Contains(s) {
				whole = false
				break
			}
		}
		if whole {
			out[cpu] = struct{}{}
		}
	}
	return out, nil
}
