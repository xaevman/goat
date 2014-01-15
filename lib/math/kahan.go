//  ---------------------------------------------------------------------------
//
//  kahan.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package math

// Stdlib imports.
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

/* function KahanSum(input)
    var sum = 0.0
    var c = 0.0                  // A running compensation for lost low-order bits.
    for i = 1 to input.length do
        var y = input[i] - c     // So far, so good: c is zero.
        var t = sum + y          // Alas, sum is big, y small, so low-order digits of y are lost.
        c = (t - sum) - y // (t - sum) recovers the high-order part of y; subtracting y recovers -(low part of y)
        sum = t           // Algebraically, c should always be zero. Beware overly-aggressive optimizing compilers!
        // Next time around, the lost low part will be added to y in a fresh attempt.
    return sum 
*/

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
