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

package lifecycle

// Stdlib imports.
import (
    "log"
    "testing"
    "time"
)

// Test objects.
var syncObj *Lifecycle

// TestLifecycleWithHeartbeat runs a work proces with heartbeat
// enabled and makes sure no blocking or race conditions occur.
func TestLifecycleWithHeartbeat(t *testing.T) {
    log.Println("TestLifecycleWithHeartbeat: startup")
    syncObj = New()

    log.Println("TestLifecycleWithHeartbeat: start heart")
    syncObj.StartHeart(1000)

    log.Println("TestLifecycleWithHeartbeat: doWork")
    go doWork()

    log.Println("TestLifecycleWithHeartbeat: wait 5 sec")
    <-time.After(5 * time.Second)

    log.Println("TestLifecycleWithHeartbeat: shutdown")
    syncObj.Shutdown()

    log.Println("TestLifecycleWithHeartbeat: passed")
}

// TestLifecycle runs a work process without heartbeat enabled
// and makes sure no blocking or race conditions occur.
func TestLifecycle(t *testing.T) {
    log.Println("TestLifecycleSync: startup")
    syncObj = New()

    log.Println("TestLifecycleSync: doWork")
    go doWork()

    log.Println("TestLifecycleSync: shutdown")
    syncObj.Shutdown()

    log.Println("TestLifecycleSync: passed")
}

// doWork is the async work process which runs in its own go routine
// during the test.
func doWork() {
    for syncObj.QueryRun() {
        select {
        case <-syncObj.QueryHeartbeat():
            log.Println("Heartbeat received")
        case <-syncObj.QueryShutdown():
        }
    }

    log.Println("doWork: shutdown complete")
    syncObj.ShutdownComplete()
}
