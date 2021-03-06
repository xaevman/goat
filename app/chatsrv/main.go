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
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/proto/dbg"
)

// Application name.
const APP_NAME = "ChatSrv"

// Default config options.
const (
    DEFAULT_DBG_ADDR = "127.0.0.1:8910"
    DEFAULT_DIAG_URI = "127.0.0.1:8911"
    DEFAULT_TCP_ADDR = "127.0.0.1:8900"
    DEFAULT_UDP_ADDR = "127.0.0.1:8901"
)

// Perf counters.
const (
    PERF_CHATSRV_CONNECT_MSG = iota
    PERF_CHATSRV_CREATE_CHAN
    PERF_CHATSRV_DIST_MSG
    PERF_CHATSRV_JOIN_CHAN
    PERF_CHATSRV_LEAVE_CHAN
    PERF_CHATSRV_SET_NAME
    PERF_CHATSRV_COUNT
)

// Perf friendly names.
var perfNames = []string {
    "ConnectMsg",
    "CreateChannel",
    "DistributeMsg",
    "JoinChannel",
    "LeaveChannel",
    "SetName",
}

// Perf object.
var perfs = perf.NewCounterSet(
    APP_NAME,
    PERF_CHATSRV_COUNT,
    perfNames,
)

// Protocol instances.
var(
    dbgProto  = net.NewProtocol(APP_NAME + "Dbg", new(dbg.DbgSrv))
    chatproto = net.NewProtocol(APP_NAME, new(ChatSrv))
)


// main is the application entry point.
func main() {
    goapp.SetAppStarter(new(ChatSrvStart))
    goapp.SetLoopHandler(new(ChatSrvLoop))

    goapp.SetHeartbeat(1000) // 1000ms / 1sec

    stopChan := goapp.Start(APP_NAME)
    <-stopChan
}
