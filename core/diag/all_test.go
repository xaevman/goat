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
    "fmt"
    "log"
    "testing"
    "time"
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


// TestDiag creates diag objects and formats them as strings and json.
// If the process doesn't crash itself, the test passes!
func TestDiag(t *testing.T) {
    perfs := perf.NewCounterSet("Module.Diag", PERF_DIAG_COUNT, perfNames)
    perfs.Increment(PERF_DIAG_TEST1)

    for i := 0; i < 2299; i++ {
        perfs.Increment(PERF_DIAG_TEST2)
    }

    diag := New()

    str := AsString(diag)
    log.Println("AsString ****************************")
    log.Println(str)

    log.Println()
    log.Println()

    json := AsJson(diag)
    log.Println("AsJson ******************************")
    log.Println(json)
}


func TestBlocked(t *testing.T) {
    fmt.Println()
    log.Println("Testing blocked output ==============")
    go func() {
        <-time.After(1 * time.Second)
        fmt.Println(NewBlockedData())
    }()
    
    <-time.After(4 * time.Second)
}
