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
func (this *ChatSrvStart) PreInit() {
	config.InitIniProvider("chat.ini", 1)
	debugLogs, _ := config.GetBoolVal("System.DebugLogs", 0, false)
	log.DebugLogs = debugLogs
}
func (this *ChatSrvStart) PostInit() {
	addr, _ := config.GetVal("Net.SrvAddr", 0, DEFAULT_ADDR)
	proto.ListenTcp(addr)
}


// ChatSrvLoop is a goapp.LoopHandler implementation for a ChatSrv
// instance.
type ChatSrvLoop struct {}
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
func (this *ChatSrvLoop) PreLoop() {}
func (this *ChatSrvLoop) PostLoop() {}
