//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package perf

import (
	"github.com/xaevman/goat/lib/math"
	"github.com/xaevman/goat/lib/str"
)

// Stdlib imports.
import (
	"fmt"
	"log"
	"math/rand"
	"testing"
)

// Perf definitions for this module.
const (
	PERF_TEST_COUNTER1 = iota
	PERF_TEST_COUNTER2
	PERF_TEST_COUNTER3
	PERF_TEST_COUNTER4
	PERF_TEST_COUNTER5
	PERF_TEST_COUNTER6
	PERF_TEST_COUNTER7
	PERF_TEST_COUNT
)

// Friendly names for PERF_TEST counters.
var perfNames = []string{
	"TestCounter1",
	"TestCounter2",
	"TestCounter3",
	"TestCounter4",
	"TestCounter5",
	"TestCounter6",
	"TestCounter7",
}

// Stat test helper object.
type statResult struct {
	min      int64
	max      int64
	mean     float64
	variance float64
	stdDev   float64
}

// Stat test data.
var statSeries = []int64 {
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

// TestPerfs test variables
var (
	goCount   int = 1000
	testCount int = 1000
)

// TestPerfs makes sure a perfs object can be saved and then retrieved from the perf system.
func TestPerfStore(t *testing.T) {
	// create object
	perfs := NewCounters(
		"test",
		PERF_TEST_COUNT,
		perfNames,
	)

	// make sure we can grab it back out of the perfs service
	p2 := GetPerfs(perfs.name)

	if perfs != p2 {
		t.Fatalf("GetPerfs returned a different object! (%+v)\n", p2)
	}

	log.Println("TestPerfStore: passed")
}

// TestPerfCounts randomly increments test counters and then checks that the end results match
// the number of iterations that were performed during the test.
func TestPerfCounts(t *testing.T) {
	// create object
	perfs := NewCounters(
		"test",
		PERF_TEST_COUNT,
		perfNames,
	)

	// spawn some go routines so that updates arrive at indeterminate
	// intervals
	doneChan := make(chan bool)
	for i := 0; i < goCount; i++ {
		go func() {
			for x := 0; x < testCount; x++ {
				perfs.Increment(rand.Intn(PERF_TEST_COUNT))
			}

			doneChan <- true
		}()
	}

	// wait for all routines to finish
	for i := 0; i < goCount; i++ {
		<-doneChan
	}

	// check totals against iterations
	var total int64 = 0
	for i := 0; i < PERF_TEST_COUNT; i++ {
		total += perfs.Get(i).Value()
	}

	if total != int64(goCount * testCount) {
		t.Fatalf(
			"Perf total != iterations (%v, %v)",
			total,
			goCount * testCount,
		)
	}

	// success!
	log.Println(perfs)
	log.Printf("Total: %v\n", total)
	log.Println("TestPerfCounts: passed")
}

// TestStats runs 120 values through a Stat object and checks the resulting
// statistics against a known-correct answer set after every 10 new values.
// Note that if STAT_SAMPLES is chaged the answer set also needs to be 
// re-computed.
func TestStats(t *testing.T) {
	fmt.Println()

	var counter, resCounter int

	s := new(Stat)

	for i := 0; i < len(statSeries); i++ {
		s.Next(statSeries[i])

		counter++
		if counter > 9 {
			min := math.IClamp(i + 1 - STAT_SAMPLES, 0, len(statSeries))
			fmt.Println()
			fmt.Printf(
				"\n========\nset %v :: %v\n",
				resCounter,
				str.Int64ArrayToList(statSeries[min:i + 1], ","),
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

			if math.Round(s.mean, 2) != math.Round(ex.mean, 2) {
				t.Fatalf(
					"Mean[%v] expected %v, result %v",
					i,
					ex.mean,
					s.mean,
				)
			}

			if math.Round(s.variance, 2) != math.Round(ex.variance, 2) {
				t.Fatalf(
					"Variance[%v] expected %v, result %v",
					i,
					ex.variance,
					s.variance,
				)
			}

			if math.Round(s.stdDev, 2) != math.Round(ex.stdDev, 2) {
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
