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
	"github.com/xaevman/goat/core/goapp"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/perf"
)

// Application name.
const APP_NAME = "ChatSrv"

// Default config options.
const (
	DEFAULT_ADDR = "127.0.0.1:8900"
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

// ChatSrv protocol instance.
var proto = net.NewProtocol(APP_NAME, new(ChatSrv))


// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ChatSrvStart))
	goapp.SetLoopHandler(new(ChatSrvLoop))

	goapp.SetHeartbeat(1000) // 1000ms / 1sec

	stopChan := goapp.Start(APP_NAME)
	<-stopChan
}
