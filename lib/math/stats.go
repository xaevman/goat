//  ---------------------------------------------------------------------------
//
//  stats.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package math includes helper objects for some commonly useful
// situations, such as improving floating point summation accuracy,
// and statistical counter objects.
package math

// Stdlib imports.
import (
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
	max       int
	maxCursor int
	mean      float64
	min       int
	mutex     sync.Mutex
	stale     bool
	stdDev    float64
	vals      [STAT_SAMPLES]int
	variance  float64
}

// MaxCount returns the total number of items in the set.
func (this *Stat) Len() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.maxCursor
}

// Max returns the largest of all values in the set.
func (this *Stat) Max() int {
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
func (this *Stat) Min() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.min
}

// NextVal adds a new value to the next slot in the set. Older values
// are overwritten in the order that they were added.
func (this *Stat) NextVal(val int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

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

// StdDev returns the standard deviation across all values in the set.
func (this *Stat) StdDev() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.stdDev
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

// recalc traverses the set of values and recomputes the min, max, mean, 
// variance and standard deviation.
func (this *Stat) recalc() {
	var total int

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
