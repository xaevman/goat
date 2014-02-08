//  ---------------------------------------------------------------------------
//
//  chatsrvapp.go
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
import(
	"github.com/xaevman/goat/core/config"
	"github.com/xaevman/goat/core/diag"
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import(
	"fmt"
)


// ChatSrvStart is a goapp.AppStarter implementation for a ChatSrv
// instance.
type ChatSrvStart struct {}

// PreInit registers an ini config provider and queries the config
// system to determine if debug logs should be enabled during this
// run.
func (this *ChatSrvStart) PreInit() {
	config.InitIniProvider("config/chat.ini", 1)
	debugLogs, _ := config.GetBoolVal("System.DebugLogs", 0, false)
	log.DebugLogs = debugLogs
}

// PostInit queries the config system to determine which bind address
// the server should listen on.
func (this *ChatSrvStart) PostInit() {
	addr, _ := config.GetVal("Net.SrvAddrTcp", 0, DEFAULT_TCP_ADDR)
	proto.ListenTcp(addr)

	addr, _ = config.GetVal("Net.SrvAddrUdp", 0, DEFAULT_UDP_ADDR)
	proto.ListenUdp(addr)

	addr, _ = config.GetVal("Debug.SrvAddr", 0, DEFAULT_DBG_ADDR)
	dbgProto.ListenTcp(addr)

	addr, _ = config.GetVal("Net.SrvAddrHttp", 0, DEFAULT_DIAG_URI)
	if addr != "" {
		net.InitHttpSrv(addr)
		diag.InitWebDiag()
	}
}


// ChatSrvLoop is a goapp.LoopHandler implementation for a ChatSrv
// instance.
type ChatSrvLoop struct {}

// OnHeartbeat queries the perf system for basic rx/tx stats and
// prints their rates to the console.
func (this *ChatSrvLoop) OnHeartbeat() {
	chatCounters := perf.GetCounterSet("Module.Net.Proto.ChatSrv")
	rx           := chatCounters.Get(net.PERF_PROTO_RCV_OK)
	tx           := chatCounters.Get(net.PERF_PROTO_SEND_OK)

	perfStr := fmt.Sprintf(
		"rx/tx sec: %d/%d",
		rx.PerSec(),
		tx.PerSec(),
	)

	log.Info(perfStr)
}

// PreLoop is unused in ChatSrvLoop.
func (this *ChatSrvLoop) PreLoop() {}

// PostLoop is unused in ChatSrvLoop.
func (this *ChatSrvLoop) PostLoop() {}
