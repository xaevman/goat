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

// External imports.
import (
	"github.com/xaevman/goat/lib/strutil"
)

// Stdlib imports.
import(
	"fmt"
	"testing"
)

// Stat test helper object.
type statResult struct {
	min      int
	max      int
	mean     float64
	variance float64
	stdDev   float64
}

// Stat test data.
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

// Stat test validated results.
var statResults = []statResult {
	statResult{ -177, 116, -10.90, 12368.10, 111.21 },
	statResult{ -198, 186, -13.05, 13768.16, 117.34 },
	statResult{ -198, 186, -35.27, 12872.96, 113.46 },
	statResult{ -198, 186, -24.50, 12278.82, 110.81 },
	statResult{ -198, 188, -21.06, 12256.42, 110.71 },
	statResult{ -198, 198,  -9.43, 13594.08, 116.59 },
	statResult{ -198, 198,  -9.69, 13398.22, 115.75 },
	statResult{ -198, 198,  -5.89, 13696.61, 117.03 },
	statResult{ -198, 198,  -9.44, 12650.56, 112.47 },
	statResult{ -198, 198,  -4.58, 11982.75, 109.47 },
	statResult{ -198, 198,  -0.44, 12057.54, 109.81 },
	statResult{ -198, 198,   4.22, 11869.63, 108.95 },
}

// KahanSum test data.
var kahanSeries = []float64 {
	-7949007.7765421,
	-5199414.0144435,
	-4291398.7088443,
	-3128598.5247831,
	2054702.9339032,
	6787782.3099437,
	2955415.9347692,
	7280850.111172,
	7461681.8863254,
	-589385.74539003,
	6361060.8672542,
	-9775664.9552731,
	-4367239.1746041,
	6086489.3980727,
	4360032.1627967,
	4508700.1354008,
	3690154.074547,
	5144006.4400174,
	-1579990.4296081,
	-8241152.4924641,
	-7459892.8901646,
	2597973.3153237,
	-6884611.0985077,
	-6886788.7076395,
	3390125.089041,
	-8274028.3190617,
	8453052.6206144,
	8240321.1753072,
	1221807.2690171,
	7212656.3439205,
	-6639193.7884685,
	-7422158.8752336,
	-5055273.9273083,
	-2153998.0136575,
	9308815.6074792,
	2929914.534525,
	-8490178.3142659,
	-5510705.0740676,
	42938.566786675,
	3241735.5260075,
	-8310228.5760968,
	770721.6640798,
	-5576605.5153574,
	717372.69438681,
	-7619913.4055618,
	-6790661.4378051,
	9236646.7226467,
	-4602194.8822784,
	-8675034.0548693,
	7203619.2460002,
}

// KahanSum test result.
var kahanResult = -3.0214742072958037e+07

// TestStats runs 120 values through a Stat object and checks the resulting
// statistics against a known-correct answer set after every 10 new values.
func TestStats(t *testing.T) {
	fmt.Println()

	var counter, resCounter int

	s := new(Stat)

	for i := 0; i < len(statSeries); i++ {
		s.NextVal(statSeries[i])

		counter++
		if counter > 9 {
			min := IClamp(i + 1 - STAT_SAMPLES, 0, len(statSeries))
			fmt.Println()
			fmt.Printf(
				"\n========\nset %v :: %v\n",
				resCounter,
				strutil.IntArrayToList(statSeries[min:i + 1], ","),
			)

			printStats(i, s)

			ex := statResults[resCounter]

			if s.min != ex.min {
				t.Fatalf(
					"Min[%v] expected %v, result %v",
					i,
					ex.min, 
					s.min,
				)
			}

			if s.max != ex.max {
				t.Fatalf(
					"Max[%v] expected %v, result %v",
					i,
					ex.max,
					s.max,
				)
			}

			if Round(s.mean, 2) != Round(ex.mean, 2) {
				t.Fatalf(
					"Mean[%v] expected %v, result %v",
					i,
					ex.mean,
					s.mean,
				)
			}

			if Round(s.variance, 2) != Round(ex.variance, 2) {
				t.Fatalf(
					"Variance[%v] expected %v, result %v",
					i,
					ex.variance,
					s.variance,
				)
			}

			if Round(s.stdDev, 2) != Round(ex.stdDev, 2) {
				t.Fatalf(
					"StdDev[%v] expected %v, result %v",
					i,
					ex.stdDev,
					s.stdDev,
				)
			}

			fmt.Printf("\nSet %v: passed\n", resCounter)

			counter = 0
			resCounter++
		}
	}

	fmt.Println()
}

// TextKahanSum adds 50 floats with a wide range of values and checks results.
func TestKahanSum(t *testing.T) {
	fmt.Println()

	ks := new(KahanSum)

	for i := 0; i < len(kahanSeries); i++ {
		ks.Add(kahanSeries[i])
	}

	if ks.Sum() != kahanResult {
		t.Fatalf("Bad sum: %v != %v", ks.Sum(), kahanResult)
	}

	fmt.Printf("With KahanSum: %v\n", ks.Sum())

	var sum float64
	for i := 0; i < len(kahanSeries); i++ {
		sum += kahanSeries[i]
	}
	fmt.Printf("Std sum:       %v\n", sum)

	fmt.Println()
}

// printStats prints a summary of the contents of a Stat object.
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
