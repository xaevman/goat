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

package goapp

// External imports.
import(
    "github.com/xaevman/goat/core/log"
)

// Stdlib imports.
import(
    "testing"
    "time"
)

// TestDefaultApp tests a very basic GoApp using all default interfaces.
// It starts the app and then shuts itself down after 10 seconds.
func TestDefaultApp(t *testing.T) {
    log.DebugLogs = true

    SetHeartbeat(1 * 1000)  // sec * ms
    go waitForShutdown()

    stopChan := Start("DefaultApp")
    <-stopChan
}

// waitForShutdown waits 10 seconds and then starts the application shutdown
// sequence.
func waitForShutdown() {
    <-time.After(10 * time.Second)
    Stop()
}
