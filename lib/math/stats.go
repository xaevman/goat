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

package math

import (
	"math"
	"sync"
)

type RunningStats struct {
	cursor   byte
	max      int
	mean     float64
	min      int
	mutex    sync.Mutex
	stale    bool
	stdDev   float64
	vals     [256]int
	variance float64
}

func NewStats(initialSample int) *RunningStats {
	stats := new(RunningStats)

	for i := 0; i < len(stats.vals); i++ {
		stats.vals[i] = initialSample
	}

	stats.NextVal(initialSample)

	return stats
}

func (this *RunningStats) Max() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.max
}

func (this *RunningStats) Mean() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}
	
	return this.mean
}

func (this *RunningStats) Min() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.min
}

func (this *RunningStats) NextVal(val int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.vals[this.cursor] = val	
	this.stale             = true
	this.cursor++
}

func (this *RunningStats) StdDev() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.stdDev
}

func (this *RunningStats) Variance() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stale {
		this.recalc()
	}

	return this.variance
}

func (this *RunningStats) recalc() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	var total int

	for i := 0; i < len(this.vals); i++ {
		total += this.vals[i]

		if this.vals[i] > this.max {
			this.max = this.vals[i]
		}

		if this.vals[i] < this.min {
			this.min = this.vals[i]
		}
	}

	this.mean = float64(total) / float64(len(this.vals))

	var diff   [256]float64
	var ftotal float64

	for i := 0; i < len(this.vals); i++ {
		diff[i] = math.Pow(float64(this.vals[i]) - this.mean, 2)
		ftotal += diff[i]
	}

	this.variance = ftotal / float64(len(this.vals))
	this.stdDev   = math.Sqrt(this.variance)

	this.stale = false
}
