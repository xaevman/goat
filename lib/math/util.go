//  ---------------------------------------------------------------------------
//
//  util.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package math

func IClamp(val int, min int, max int) int {
	if val < min {
		return min
	}

	if val > max {
		return max
	}

	return val
}
