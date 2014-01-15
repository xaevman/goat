//  ---------------------------------------------------------------------------
//
//  perfcounters.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package perf

// Stdlib imports.
import (
	"bytes"
	"fmt"
	"sync/atomic"
)


// PerfCounters represent a collection of named performance counters and
// their values.
type PerfCounters struct {
	name    string
	nameMap map[int]string
	perfs   []uint64
}

// NewPerfCounters is a construction helper which creates a new PerfCounters
// container, registers and initializes it, and returns a pointer to it for use.
func NewPerfCounters(name string, size int, names []string) *PerfCounters {
	if len(names) != size {
		return nil
	}

	counters := PerfCounters {
		nameMap: make(map[int]string, size),
		perfs:   make([]uint64, size),
	}

	for i := range names {
		counters.nameMap[i] = names[i]
	}

	registerPerfs(&counters)

	return &counters
}

// Get returns the current value of the specified counter, or returns 0
// if no counter exists at that offset.
func (this *PerfCounters) Get(offset int) uint64 {
	if offset > len(this.perfs) {
		return 0
	}

	return atomic.LoadUint64(&this.perfs[offset])
}

// Increment adds 1 to the value of the specified counter. If the specified
// offset doesn't exist, it's a no-op.
func (this *PerfCounters) Increment(offset int) {
	if offset > len(this.perfs) {
		return
	}

	atomic.AddUint64(&this.perfs[offset], 1)
}

// Name returns the friendly name of this PerfCounters container.
func (this *PerfCounters) Name() string {
	return this.name
}

// Reset ranges through all individual perfs in this PerfCounters object and
// sets them all back to zero.
func (this *PerfCounters) Reset() {
	for i := range this.perfs {
		atomic.StoreUint64(&this.perfs[i], 0)
	}
}

// String returns a friendly representation of all the values stored in this
// PerfCounters object.
func (this *PerfCounters) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Perf dump")

	for i := range this.perfs {
		buffer.WriteString(fmt.Sprintf(
			"\n%v = %v",
			this.nameMap[i],
			this.perfs[i],
		))
	}

	return buffer.String()
}
