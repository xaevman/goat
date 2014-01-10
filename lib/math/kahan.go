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

//function KahanSum(input)
//    var sum = 0.0
//    var c = 0.0                  	// A running compensation for lost low-order bits.
//    for i = 1 to input.length do
//        var y = input[i] - c     	// So far, so good: c is zero.
//        var t = sum + y          	// Alas, sum is big, y small, so low-order digits of y are lost.
//        c = (t - sum) - y 			// (t - sum) recovers the high-order part of y; subtracting y recovers -(low part of y)
//        sum = t           			// Algebraically, c should always be zero. Beware overly-aggressive optimizing compilers!
//    								// Next time around, the lost low part will be added to y in a fresh attempt.
//    return sum
type KahanSum struct {
	compensation float64
	sum        	 float64
}

func (this *KahanSum) Add(val float64) float64 {
	y := val - this.compensation
	t := this.sum + y
	this.compensation = (t - this.sum) - y
	this.sum = t

	return this.sum
}

func (this *KahanSum) Reset() {
	this.compensation = 0
	this.sum          = 0
}

func (this *KahanSum) Sum() float64 {
	return this.sum
}
