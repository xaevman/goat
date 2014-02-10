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
    "encoding/json"
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

// TestStacks prints the object-based stack produced by NewStackTrace, and
// compares it to the string-formatted output produced by NewStackString.
func TestStacks(t *testing.T) {
    go func() {
        for {
            <-time.After(1 * time.Second)
        }
    }()

    go func() {
        for {
            <-time.After(1 * time.Second)
        }
    }()

    results := NewStackTrace()
    if results == nil {
        t.Fatalf("Stack trace nil")
    }

    json, _ := json.MarshalIndent(results, "", "    ")    
    fmt.Println(string(json))

    fmt.Println()

    fmt.Println(NewStackString())
}

// TestDiag creates diag objects and formats them as strings and json.
// If the process doesn't crash itself, the test passes!
func _TestDiag(t *testing.T) {
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

// TestBlocked creates diagnostic information for blocked goroutines
// and prints it to the console for validation.
func _TestBlocked(t *testing.T) {
    fmt.Println()
    log.Println("Testing blocked output ==============")
    go func() {
        <-time.After(1 * time.Second)
        fmt.Println(NewBlockedData())
    }()
    
    <-time.After(4 * time.Second)
}
