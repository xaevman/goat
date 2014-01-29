//  ---------------------------------------------------------------------------
//
//  stat.go
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
	"fmt"
	"math"
	"sync"
)

// The number of samples stored and used to calculate statistics
// in a given Stat object.
const STAT_SAMPLES = 100


// Stat represents a series of values and some common statistical values
// associated with that set.
type Stat struct {
	cursor    int
	max       int64
	maxCursor int
	mean      float64
	min       int64
	mutex     sync.Mutex
	stale     bool
	stdDev    float64
	vals      [STAT_SAMPLES]int64
	variance  float64
}

// NewStat initializes a new Stat object and returns a pointer to it for use.
func NewStat() *Stat {
	newStat := new(Stat)
	newStat.Reset()

	return newStat
}

// MaxCount returns the total number of items in the set.
func (this *Stat) Len() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.maxCursor
}

// Max returns the maximum value that has been observed since recording
// stats for this object.
func (this *Stat) Max() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.max
}

// Mean returns the mean (simple average) across all values in the set.
func (this *Stat) Mean() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}
	
	return this.mean
}

// Min returns the minimum value that has been observed since reocrding
// stats for this object.
func (this *Stat) Min() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.min
}

// Next submits the given value as the next value in the set.
func (this *Stat) Next(val int64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.nextVal(val)
}

// Reset re-initializes all stat values back to zero.
func (this *Stat) Reset() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.cursor    = 0
	this.max       = 0
	this.maxCursor = 0
	this.mean      = 0
	this.min       = math.MaxInt64
	this.stale     = false
	this.stdDev    = 0
	this.variance  = 0
}

// StdDev returns the standard deviation across all values in the set.
func (this *Stat) StdDev() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.stdDev
}

// String implements Stringer to pretty-print the Stat object.
func (this *Stat) String() string {
	return fmt.Sprintf(
		"min: %d, max: %d, mean: %.2f, " +
		"variance: %.2f, stdDev: %.2f, samples: %d",
		this.Min(),
		this.Max(),
		this.Mean(),
		this.Variance(),
		this.StdDev(),
		this.Len(),
	)
}

// Variance returns the variance across all values in the set.
func (this *Stat) Variance() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.variance
}

// nextVal adds a new value to the next slot in the set. Older values
// are overwritten in the order that they were added.
func (this *Stat) nextVal(val int64) {
	this.vals[this.cursor] = val	
	this.stale             = true

	if val < this.min {
		this.min = val
	}

	if val > this.max {
		this.max = val
	}

	if this.cursor < len(this.vals) - 1 {
		this.cursor++
	} else {
		this.cursor = 0
	}

	if this.maxCursor < len(this.vals) {
		this.maxCursor++
	}
}

// recalc traverses the set of values and recomputes the min, max, mean, 
// variance and standard deviation.
func (this *Stat) recalc() {
	var total int64

	for i := 0; i < this.maxCursor; i++ {
		total += this.vals[i]
	}

	this.mean = float64(total) / float64(this.maxCursor)

	var diff   [STAT_SAMPLES]float64
	var ftotal float64

	for i := 0; i < this.maxCursor; i++ {
		diff[i] = math.Pow(float64(this.vals[i]) - this.mean, 2)
		ftotal += diff[i]
	}

	this.variance = ftotal / float64(this.maxCursor - 1)
	this.stdDev   = math.Sqrt(this.variance)

	this.stale = false
}
