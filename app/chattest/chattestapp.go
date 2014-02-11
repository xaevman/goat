//  ---------------------------------------------------------------------------
//
//  main.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package main

// External imports.
import (
    "github.com/xaevman/goat/mod/goapp"
    "github.com/xaevman/goat/mod/log"
)

// Stdlib imports.
import (
    "os"
    "strconv"
)


// ChatTestStarter is a goapp.AppStarter implementation for a
// ChatTest instance.
type ChatTestStarter struct {}

// PreInit parses command line arguments for the target server
// address, test index, and test count.
func (this *ChatTestStarter) PreInit() {
    if len(os.Args) > 1 {
        srvAddr = os.Args[1]
    }
    if len(os.Args) > 2 {
        myIndex, _ = strconv.Atoi(os.Args[2])
        myName     = "chattest" + os.Args[2]
    }
    if len(os.Args) > 3 {
        maxTests, _ = strconv.Atoi(os.Args[3])
    }

    log.Info(
        "Startup params (addr: %s, index: %d, tests: %d)",
        srvAddr,
        myIndex,
        maxTests,
    )
}

// PostInit attempts to open a connection with the server, 
// shutting down the application if one can't be established.
func (this *ChatTestStarter) PostInit() {
    err := protocol.DialTcp(srvAddr)
    if err != nil {
        goapp.Stop()
    }
}


// ChatTestCloser is a goapp.AppCloser implementation for a
// ChatTest instance.
type ChatTestCloser struct {}

// PreShutdown performs no actions in ChatTest.
func (this *ChatTestCloser) PreShutdown() {}

// PostShutdown retreives the run statistics from the event
// handler, logs them, and sets the exit code appropriately.
func (this *ChatTestCloser) PostShutdown() {
    success, errors, exeTime := evtHandler.GetResults()

    log.Info(
        "TEST RESULTS (Success: %d, Failure: %d)",
        success,
        errors,
    )
    log.Info(
        "TEST RESULTS (Execution time: %v)",
        exeTime,
    )
    log.Info(
        "TEST RESULTS (%.2f msg/sec)",
        float64(maxTests) / exeTime.Seconds(),
    )

    if success == maxTests {
        goapp.SetExitCode(0)
    }

    goapp.SetExitCode(errors)   
}
