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
const APP_NAME = "ChatCli"

// Config options.
var (
    srvAddr  = "127.0.0.1:8900"
    userName = "Anon"
    useUdp   = false
)

// ChatCli protocol instance and event handler.
var (
    chatCli   = new(ChatCli)
    chatproto = net.NewProtocol(APP_NAME, chatCli)
)

// main is the application entry point.
func main() {
    goapp.SetAppStarter(new(ConsoleCliStart))

    stopChan := goapp.Start(APP_NAME)
    <-stopChan
}
