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
	lastVal   int64
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

// Increment calculates and inserts a new NextVal() which is 1 greater
// than the last value.
func (this *Stat) Increment() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.nextVal(this.lastVal + 1)
}

// MaxCount returns the total number of items in the set.
func (this *Stat) Len() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.maxCursor
}

// Max returns the largest of all values in the set.
func (this *Stat) Max() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

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

// Min returns the smallest of all values in the set.
func (this *Stat) Min() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.min
}

// Next submits the given value as the next value in the set.
func (this *Stat) Next(val int64) {
	this.nextVal(val)
}

// Reset re-initializes all stat values back to zero.
func (this *Stat) Reset() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.cursor    = 0
	this.lastVal   = 0
	this.max       = 0
	this.maxCursor = 0
	this.mean      = 0
	this.min       = 0
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
	this.mutex.Lock()
	defer this.mutex.Unlock()
	
	return fmt.Sprintf("%d", this.lastVal)
}

// Value returns the most recent value submitted to the Stat object.
func (this *Stat) Value() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.lastVal
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
	this.lastVal           = val
	this.vals[this.cursor] = val	
	this.stale             = true

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

		if this.vals[i] > this.max {
			this.max = this.vals[i]
		}

		if this.vals[i] < this.min {
			this.min = this.vals[i]
		}
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
