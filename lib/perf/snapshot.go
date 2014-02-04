//  ---------------------------------------------------------------------------
//
//  snapshot.go
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
	"math"
	"time"
)


// TakeSnapshot creates a new Snapshot object containing the values
// of all the current metrics in the perf system, and returns a 
// pointer to it for use.
func TakeSnapshot() *Snapshot {
	snap          := new(Snapshot)
	snap.Counters  = make([]*CounterVals, 0)
	snap.Timestamp = time.Now()

	counterSets := GetAllCounterSets()
	for _, counterSet := range counterSets {
		count := counterSet.Len()
		for i := 0; i < count; i++ {
			counter          := counterSet.Get(i)
			newVals          := new(CounterVals)
			newVals.Name      = counterSet.CounterName(i)
			newVals.Value     = counter.Value()
			newVals.PerSec    = counter.PerSec()
			newVals.MaxPerSec = counter.MaxPerSec()

			snap.Counters = append(snap.Counters, newVals)

			stats := counter.Stats()
			if stats == nil {
				continue
			}

			newVals.Max      = stats.Max()
			newVals.Mean     = stats.Mean()
			newVals.Min      = stats.Min()
			newVals.StdDev   = stats.StdDev()
			newVals.Variance = stats.Variance()
		}
	}

	return snap
}


// CounterVals represents all of the values associated with a given 
// counter.
type CounterVals struct {
	Max       int64
	MaxPerSec int64
	Mean      float64
	Min       int64
	Name      string
	PerSec    int64
	StdDev    float64
	Value     int64
	Variance  float64
}

// String pretty-prints the values inside the CounterVals object.
func (this *CounterVals) String() string {
	return fmt.Sprintf(
		"%s val:%d /sec:%d max/Sec:%d min:%d " +
		"max:%d mean:%.2f variance:%.2f stddev:%.2f\n",
		this.Name,
		this.Value,
		this.PerSec,
		this.MaxPerSec,
		this.Min,
		this.Max,
		this.Mean,
		this.Variance,
		this.StdDev,
	)
}

// StringBrief prints a less verbose output of performance counters, only
// writing data for non-default values.
func (this *CounterVals) StringBrief() string {
	var buffer bytes.Buffer

	if this.Value != 0 {
		buffer.WriteString(fmt.Sprintf("val:%d ", this.Value))
	}

	if this.PerSec != 0 {
		buffer.WriteString(fmt.Sprintf("/sec:%d ", this.PerSec))
	}

	if this.MaxPerSec != 0 {
		buffer.WriteString(fmt.Sprintf("max/sec:%d ", this.MaxPerSec))
	}

	if this.Min != math.MaxInt64 && this.Min != 0 {
		buffer.WriteString(fmt.Sprintf("min:%d", this.Min))
	}

	if this.Max != 0 {
		buffer.WriteString(fmt.Sprintf("max:%d ", this.Max))
	}

	if this.Mean != 0 {
		buffer.WriteString(fmt.Sprintf("mean:%.2f ", this.Mean))
	}

	if this.Variance != 0 {
		buffer.WriteString(fmt.Sprintf("variance:%.2f ", this.Variance))
	}

	if this.StdDev != 0 {
		buffer.WriteString(fmt.Sprintf("stddev:%.2f", this.StdDev))
	}

	counterStr := buffer.String()
	if len(counterStr) > 0 {
		return fmt.Sprintf("%s %s\n", this.Name, counterStr)
	}

	return ""
}

// Snapshot represents a snapshot of perf counter data taken at a given
// point in time.
type Snapshot struct {
	Counters  []*CounterVals
	Timestamp time.Time
}

// String pretty-prints the Snapshot object and all of the counter objects
// it contains.
func (this *Snapshot) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Timestamp: %v\n", this.Timestamp))
	for _, counter := range this.Counters {
		buffer.WriteString(counter.String())
	}

	return buffer.String()
}

// StringBrief pretty-prints a less verbose version of the Snapshot object,
// opting to only print non-default values.
func (this *Snapshot) StringBrief() string {
	var buffer bytes.Buffer

	for _, counter := range this.Counters {
		buffer.WriteString(counter.StringBrief())
	}

	return buffer.String()
}
