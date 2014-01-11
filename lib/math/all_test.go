//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package math

import "github.com/xaevman/goat/lib/strutil"

import(
	"fmt"
	"testing"
)

var statSeries = []int {
	-142,    7,  105,   87, -177,   70,   16, -148,
	 -43,  116,   84,   44,  161, -198,  -39,  -86,
	 186, -153,  -70,  -81, -157,   58, -138, -195,
	-124,   14,    2, -145,   38, -150,   48,  168,
	  11,  120,  -44,  -67,  -90,  121,  -85, -104,
	 -22,  188,  -15,  113,  115, -102,  -97,  -98,
	-168,   13,   76,  198,  185,  -36, -135,  139,
	-111,  100, -109,  180,   34,  184,  -71,  -93,
	  81,  140,  -79,  -21, -122, -165, -171,  141,
	 128, -110, -181,   64,  158,  107,   49,   22,
	 -29,   -8,  -26,  -68,   -2, -105, -161,   60,
	   8,  -48,   36,   91,    0,   50,   42,   69,
	   3,  -99,   45,  155,  109,   55, -164,   57,
	  74, -121,   61,  -51,   89,  196, -101,   92,
	  93,  -23, -112,  -54,  176,  137,  -83,  189,
}

func TestPrint(t *testing.T) {
	var counter int

	for i := 0; i < len(statSeries) ; i++ {
		counter++
		if counter > 9 {
			min := IClamp(i + 1 - STAT_SAMPLES, 0, len(statSeries))
			fmt.Println()
			fmt.Println(strutil.IntArrayToList(statSeries[min:i + 1], ","))
			counter = 0
		}
	}
}

func TestStats(t *testing.T) {
	var counter int

	s := new(Stat)

	for i := 0; i < len(statSeries) ; i++ {
		s.NextVal(statSeries[i])

		counter++
		if counter > 9 {
			printStats(i, s)
			counter = 0
		}
	}
}

func printStats(index int, s *Stat) {
	fmt.Printf(
		"\nIndex:    %v\n" +
		"Len:      %v\n" +
		"Min:      %v\n" +
		"Max:      %v\n" +
		"Mean:     %v\n" +
		"Variance: %v\n" +
		"StdDev:   %v\n",
		index,
		s.Len(),
		s.Min(),
		s.Max(),
		s.Mean(),
		s.Variance(),
		s.StdDev(),
	)
}
