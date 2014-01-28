//  ---------------------------------------------------------------------------
//
//  counters.go
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
)


// Counters represent a collection of named performance counters and
// their values.
type Counters struct {
	counters []*Stat
	name     string
	nameMap  map[int]string
}

// NewCounters is a construction helper which creates a new Counters
// container, registers and initializes it, and returns a pointer to it for use.
func NewCounters(name string, size int, names []string) *Counters {
	if len(names) != size {
		return nil
	}

	newPerfs := Counters {
		nameMap:  make(map[int]string, size),
		counters: make([]*Stat, size),
	}

	for i := range names {
		newPerfs.nameMap[i]  = names[i]
		newPerfs.counters[i] = new(Stat)
	}

	registerPerfs(&newPerfs)

	return &newPerfs
}

// Get returns the current value of the specified counter, or returns 0
// if no counter exists at that offset.
func (this *Counters) Get(offset int) *Stat {
	if offset > len(this.counters) {
		return nil
	}

	return this.counters[offset]
}

// Increment adds 1 to the value of the specified counter. If the specified
// offset doesn't exist, it's a no-op.
func (this *Counters) Increment(offset int) {
	if offset > len(this.counters) {
		return
	}

	this.counters[offset].Increment()
}

// Name returns the friendly name of this Counters container.
func (this *Counters) Name() string {
	return this.name
}

// Next sets the next value in the Stat object at the given offset.
func (this *Counters) Next(offset int, value int64) {
	if offset > len(this.counters) {
		return
	}

	this.counters[offset].Next(value)
}

// Reset ranges through all individual counters in this Counters object and
// sets them all back to zero.
func (this *Counters) Reset() {
	for i := range this.counters {
		this.counters[i].Reset()
	}
}

// String returns a friendly representation of all the values stored in this
// Counters object.
func (this *Counters) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Perf dump")

	for i := range this.counters {
		buffer.WriteString(fmt.Sprintf(
			"\n%v = %v",
			this.nameMap[i],
			this.counters[i],
		))
	}

	return buffer.String()
}
