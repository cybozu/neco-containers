package main

import "fmt"

// Lender is a pinned pod that may lend its exclusive CPUs.
type Lender struct {
	UID     string
	Lending bool // true while the lendWhile condition holds (role=replica)
}

// ComputeLendable returns the set of CPUs that may currently be lent to
// borrowers: the union of the exclusive CPUs of the named container of all
// lending lenders (containerName "" means every container of the pod),
// optionally reduced to whole physical cores.
//
// Only the designated workload container donates its CPUs: a future sidecar
// with an integer CPU request must keep its exclusivity even inside a
// lending pod. It is a pure function of the kubelet state and the lender
// list; missing state entries (e.g. a lender pod admitted but not yet
// pinned) simply contribute nothing.
func ComputeLendable(st *PinnedState, lenders []Lender, containerName string, wholeCoresOnly bool, topo *Topology) (CPUSet, error) {
	lendable := CPUSet{}
	if st == nil || !st.Static {
		// Not static (or unknown): the premise of pinning is gone. Fail
		// closed by lending nothing; the caller reports inert mode.
		return lendable, nil
	}
	for _, l := range lenders {
		if !l.Lending {
			continue
		}
		for name, reserved := range st.Reserved[l.UID] {
			if containerName != "" && name != containerName {
				continue
			}
			lendable = lendable.Union(reserved)
		}
	}
	if wholeCoresOnly && len(lendable) > 0 {
		// A CPU is lendable only when the borrower can use its whole
		// physical core: every sibling must be lent too or already in the
		// shared pool. This excludes exactly the CPUs whose sibling belongs
		// to another pod's exclusive assignment.
		filtered, err := topo.WholeCores(lendable, lendable.Union(st.SharedPool))
		if err != nil {
			return nil, fmt.Errorf("failed to filter lendable CPUs to whole cores: %w", err)
		}
		lendable = filtered
	}
	return lendable, nil
}

// DesiredBorrowerCPUSet returns the cpuset a borrower container should have:
// the kubelet-intended baseline (shared pool) plus the currently lendable
// CPUs. When lending stops this degenerates to the baseline, which is
// exactly what the kubelet itself would write.
func DesiredBorrowerCPUSet(baseline, lendable CPUSet) CPUSet {
	return baseline.Union(lendable)
}
