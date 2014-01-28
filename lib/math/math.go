//  ---------------------------------------------------------------------------
//
//  math.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package math includes helper objects for common math functions.
package math

// Stdlib imports.
import (
	"math"
)

// IClamp clamps a given int value to be between the supplied
// min and max (both inclusive).
func IClamp(val int, min int, max int) int {
	if val < min {
		return min
	}

	if val > max {
		return max
	}

	return val
}

// Round rounds numbers to the specified precision.
func Round(x float64, prec int) float64 {
	pow := math.Pow(10, float64(prec))
	
	x = x * pow

	if x < 0.0 {
		x -= 0.5
	} else {
		x += 0.5
	}

	x = float64(int64(x))

	return x / float64(pow)
}