//  ---------------------------------------------------------------------------
//
//  counterset.go
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


// CounterSet represent a collection of named performance CounterSet and
// their values.
type CounterSet struct {
	counters []*Counter
	name     string
	nameMap  map[int]string
}

// NewCounterSet is a construction helper which creates a new CounterSet
// container, registers and initializes it, and returns a pointer to it for use.
func NewCounterSet(name string, size int, names []string) *CounterSet {
	if len(names) != size {
		panic("Perf enum length must == name map length")
	}

	newCounterSet := CounterSet {
		name     : name,
		nameMap  : make(map[int]string, size),
		counters : make([]*Counter, size),
	}

	for i := range names {
		newCounterSet.nameMap[i]  = names[i]
		newCounterSet.counters[i] = NewCounter()
	}

	registerCounterSet(&newCounterSet)

	return &newCounterSet
}

// Add adds the supplied val to the value of the specified counter.
// If the specified offset doesn't exist, it's a no-op.
func (this *CounterSet) Add(offset int, val int64) {
	if offset > len(this.counters) {
		return
	}

	this.counters[offset].Add(val)
}

// CounterName returns the friendly name of the counter at the given offset.
// If the specified offset doesn't exist, a blank string is returned.
func (this *CounterSet) CounterName(offset int) string {
	if offset > len(this.counters) {
		return ""
	}

	return fmt.Sprintf("%s.%s", this.name, this.nameMap[offset])
}

// EnableStats enables statistics gathering on the given counter object.
func (this *CounterSet) EnableStats(offset int) {
	if  offset > len(this.counters) {
		return
	}

	this.counters[offset].EnableStats()
}

// Get returns the Counter object representing the given offset.
func (this *CounterSet) Get(offset int) *Counter {
	if offset > len(this.counters) {
		return nil
	}

	return this.counters[offset]
}

// Increment adds 1 to the value of the specified counter. If the specified
// offset doesn't exist, it's a no-op.
func (this *CounterSet) Increment(offset int) {
	if offset > len(this.counters) {
		return
	}

	this.counters[offset].Increment()
}

// Len returns the number of counters in available in this CounterSet.
func (this *CounterSet) Len() int {
	return len(this.counters)
}

// Name returns the friendly name of this CounterSet container.
func (this *CounterSet) Name() string {
	return this.name
}

// Reset ranges through all individual CounterSet in this CounterSet object and
// sets them all back to zero.
func (this *CounterSet) Reset() {
	for i := range this.counters {
		this.counters[i].Reset()
	}
}

// Set sets the value of the counter at the given offset to the specified value. 
// If the specified offset doesn't exist, it's a no-op
func (this *CounterSet) Set(offset int, value int64) {
	if offset > len(this.counters) {
		return
	}

	this.counters[offset].Set(value)
}

// String returns a friendly representation of all the values stored in this
// CounterSet object.
func (this *CounterSet) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Perf dump")

	for i := range this.counters {
		buffer.WriteString(fmt.Sprintf(
			"\n%s = %s",
			this.CounterName(i),
			this.counters[i].String(),
		))
	}

	return buffer.String()
}

// Value returns the current value of the counter at the given offset. If the offset
// doesn't exist, zero is returned.
func (this *CounterSet) Value(offset int) int64 {
	if offset > len(this.counters) {
		return 0
	}

	return this.counters[offset].Value()
}
