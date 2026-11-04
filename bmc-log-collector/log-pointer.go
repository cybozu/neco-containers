package main

import (
	"sync"
)

type LastPointer struct {
	LastReadId         int    // BMC SEL Id
	LastError          string // for TCP Error
	LastHttpStatusCode int    // for HTTP Error
	FirstCreateTime    int64
	// SelRead is true after SEL of the machine has been read once by this process.
	SelRead bool
}

// pointerStore keeps the pointers of each machine in memory.
// The pointers are not persisted. After restart, the collector outputs only the entries created
// within the startup lookback, so some entries may be output twice or lost.
type pointerStore struct {
	mu       sync.Mutex
	pointers map[string]LastPointer
}

func newPointerStore() *pointerStore {
	return &pointerStore{pointers: make(map[string]LastPointer)}
}

// get returns the pointer of the serial, or the zero value if there is none.
func (s *pointerStore) get(serial string) LastPointer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pointers[serial]
}

func (s *pointerStore) put(serial string, ptr LastPointer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pointers[serial] = ptr
}

// deleteDisappeared removes the pointers and metrics of the machines that disappeared from machinesList.
func (s *pointerStore) deleteDisappeared(machinesList []Machine) {
	s.mu.Lock()
	defer s.mu.Unlock()

	machineExist := make(map[string]bool, len(machinesList))
	for _, m := range machinesList {
		machineExist[m.Serial] = true
	}
	for serial := range s.pointers {
		if !machineExist[serial] {
			delete(s.pointers, serial)
			deleteMetrics(serial)
		}
	}
}
