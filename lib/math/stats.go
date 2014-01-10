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

type RunningStats struct {
	avg    float32
	cursor byte
	min    int
	max    int
	stdDev float32
	vals   [256]int
}

func NewStats(initialSample int) *RunningStats {
	stats := new(RunningStats)

	for i := 0; i < len(stats.vals); i++ {
		stats.vals[i] = initialSample
	}

	stats.NextVal(initialSample)

	return stats
}

func (this *RunningStats) Average() float32 {
	return this.avg
}

func (this *RunningStats) Max() int {
	return this.max
}

func (this *RunningStats) Min() int {
	return this.min
}

func (this *RunningStats) NextVal(val int) {
	this.vals[this.cursor] = val

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

	this.avg = float32(total) / float32(len(this.vals))
	
	this.cursor++
}

func (this *RunningStats) StdDev() float32 {
	return this.stdDev
}
