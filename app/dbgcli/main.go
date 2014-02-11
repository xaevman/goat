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

// External imports
import (
    "github.com/xaevman/goat/mod/goapp"
    "github.com/xaevman/goat/mod/net"
)

// Application name.
const APP_NAME = "DbgCli"

// Application entry point.
func main() {
    evtHandler := new(DbgCli)
    proto      := net.NewProtocol(APP_NAME, evtHandler)
    proto.DialTcp("127.0.0.1:8910")

    stopChan := goapp.Start(APP_NAME)
    <-stopChan
}
