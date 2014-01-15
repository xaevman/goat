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

// Stdlib imports.
import (
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
var perfNames = []string {
	"TestCounter1",
	"TestCounter2",
	"TestCounter3",
	"TestCounter4",
	"TestCounter5",
	"TestCounter6",
	"TestCounter7",
}

// TestPerfs test variables
var (
	goCount   int = 1000
	testCount int = 1000
)


func TestPerfs(t *testing.T) {
	// create object
	perfs := NewPerfCounters(
		"test",
		PERF_TEST_COUNT,
		perfNames,
	)

	// makes sure we can grab it back out of the perfs service
	p2 := GetPerfs(perfs.name)

	if perfs != p2 {
		t.Fatalf("GetPerfs returned a different object! (%+v)\n", p2)
	}

	// spawn some go routines so that updates arrive indeterminate
	// intervals
	doneChan := make(chan bool)
	for i := 0; i < goCount; i++ {
		go func() {
			for x := 0; x < testCount; x++ {
				perfs.Increment(rand.Intn(PERF_TEST_COUNT))
			}

			doneChan<- true
		}()
	}

	// wait for all routines to finish
	for i := 0; i < goCount; i++ {
		<-doneChan
	}

	// check totals against iterations
	var total uint64 = 0
	for i := 0; i < PERF_TEST_COUNT; i++ {
		total += perfs.Get(i)
	}

	if total != uint64(goCount * testCount) {
		t.Fatalf(
			"Perf total != iterations (%v, %v)",
			total,
			goCount * testCount,
		)
	}

	// success!
	log.Println(perfs)
	log.Printf("Total: %v\n", total)
}


