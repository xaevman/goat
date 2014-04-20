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

// GoGraph is a command line utility which can be run interactively or
// as a daemon. GoGraph queries various system APIs for configured 
// performance and statistical information about a system and then
// formats and sends the data to a Graphite server.
package main

// External imports.
import (
    "github.com/xaevman/goat/mod/goapp"
)

// Stdlib imports.
import (
    "net"
)

// Application name.
const APP_NAME  = "GoGraph"

// Graphite message format.
const GRAPH_MSG_FORMAT = "%s.%s.%s %s %d\n"

// Application config variables.
var (
    heartbeatMs = 60000
    srvAddr     = "127.0.0.1:2003"
    statPrefix  = "srv"
    sysCtlStats = make([]string, 0)
)

// Graphite server connection.
var srvCon net.Conn

// main is the application entry point.
func main() {
    goapp.SetAppStarter(new(GoGraphStart))
    goapp.SetLoopHandler(new(GoGraphLoop))
    goapp.SetCrashHandler(new(GoGraphCrash))

    stopChan := goapp.Start(APP_NAME)
    <-stopChan

    srvCon.Close()
}
