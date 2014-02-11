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
    "github.com/xaevman/goat/mod/net"
)

// Application name.
const APP_NAME = "ChatTest"

// Configurable settings.
var (
    maxTests = 100
    myIndex  = 0
    myName   = "Anon"
    srvAddr  = "127.0.0.1:8900"
)

// Network protocol and event handler.
var (
    evtHandler = new(ChatTest)
    protocol   = net.NewProtocol(APP_NAME, evtHandler)
)

// main is the application entry point.
func main() {
    goapp.SetAppStarter(new(ChatTestStarter))
    goapp.SetAppCloser(new(ChatTestCloser))

    stopChan := goapp.Start(APP_NAME)
    <-stopChan
}
