package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CPUSet is a set of logical CPU IDs.
type CPUSet map[int]struct{}

// ParseCPUSet parses a kernel cpuset list format string such as "0-2,5,7-8".
// An empty string yields an empty set.
func ParseCPUSet(s string) (CPUSet, error) {
	set := CPUSet{}
	s = strings.TrimSpace(s)
	if s == "" {
		return set, nil
	}
	for _, part := range strings.Split(s, ",") {
		lo, hi, ok := strings.Cut(part, "-")
		start, err := strconv.Atoi(strings.TrimSpace(lo))
		if err != nil {
			return nil, fmt.Errorf("invalid cpuset %q: %w", s, err)
		}
		end := start
		if ok {
			end, err = strconv.Atoi(strings.TrimSpace(hi))
			if err != nil {
				return nil, fmt.Errorf("invalid cpuset %q: %w", s, err)
			}
		}
		if end < start {
			return nil, fmt.Errorf("invalid cpuset %q: descending range %s", s, part)
		}
		for i := start; i <= end; i++ {
			set[i] = struct{}{}
		}
	}
	return set, nil
}

// String formats the set in canonical kernel list format ("0-2,5").
func (c CPUSet) String() string {
	cpus := c.Sorted()
	var b strings.Builder
	for i := 0; i < len(cpus); {
		j := i
		for j+1 < len(cpus) && cpus[j+1] == cpus[j]+1 {
			j++
		}
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		if i == j {
			fmt.Fprintf(&b, "%d", cpus[i])
		} else {
			fmt.Fprintf(&b, "%d-%d", cpus[i], cpus[j])
		}
		i = j + 1
	}
	return b.String()
}

// Sorted returns the CPU IDs in ascending order.
func (c CPUSet) Sorted() []int {
	cpus := make([]int, 0, len(c))
	for cpu := range c {
		cpus = append(cpus, cpu)
	}
	sort.Ints(cpus)
	return cpus
}

// Union returns a new set containing members of c and others.
func (c CPUSet) Union(others ...CPUSet) CPUSet {
	out := CPUSet{}
	for cpu := range c {
		out[cpu] = struct{}{}
	}
	for _, o := range others {
		for cpu := range o {
			out[cpu] = struct{}{}
		}
	}
	return out
}

// Contains reports whether cpu is a member of c.
func (c CPUSet) Contains(cpu int) bool {
	_, ok := c[cpu]
	return ok
}

// Equal reports whether c and o have the same members.
func (c CPUSet) Equal(o CPUSet) bool {
	if len(c) != len(o) {
		return false
	}
	for cpu := range c {
		if !o.Contains(cpu) {
			return false
		}
	}
	return true
}
