//  ---------------------------------------------------------------------------
//
//  kahan.go
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
	"sync"
)

// KahanSum represents a running summation of a series of floating point values.
// The Kahan sum algorithm attempts to reduce the amount of error accumulated when
// performing many math operations on floating point numbers of varying precisions.
type KahanSum struct {
	compensation float64
	mutex        sync.Mutex
	sum        	 float64
}

// Add sums the previous and supplied values together using Kahan's sum algorithm
// and returns the resulting new sum.
func (this *KahanSum) Add(val float64) float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	y := val - this.compensation
	t := this.sum + y
	this.compensation = (t - this.sum) - y
	this.sum = t

	return this.sum
}

// Reset sets the compensation and sum values back to defaults (0).
func (this *KahanSum) Reset() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.compensation = 0
	this.sum          = 0
}

// Sum returns the current sum.
func (this *KahanSum) Sum() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.sum
}
