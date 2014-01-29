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

package diag

// External imports.
import (
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	"errors"
	"log"
	"testing"
)

// Some perfs.
const (
	PERF_DIAG_TEST1 = iota
	PERF_DIAG_TEST2
	PERF_DIAG_COUNT
)
var perfNames = []string {
	"Test1",
	"Test2",
}

var e = errors.New("This is a test error")

// TestDiag creates diag objects and formats them as strings and json.
// If the process doesn't crash itself, the test passes!
func TestDiag(t *testing.T) {
	perfs := perf.NewCounterSet("Service.Diag", PERF_DIAG_COUNT, perfNames)
	perfs.Increment(PERF_DIAG_TEST1)

	for i := 0; i < 2299; i++ {
		perfs.Increment(PERF_DIAG_TEST2)
	}

	str := AsString(e)
	log.Println("AsString ****************************")
	log.Println(str)

	log.Println()
	log.Println()

	json := AsJson(e)
	log.Println("AsJson ******************************")
	log.Println(json)
}
