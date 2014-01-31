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
	"github.com/xaevman/goat/core/config"
	"github.com/xaevman/goat/core/goapp"
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/prod/chat"
)

// Stdlib imports.
import (
	"fmt"
)

// Default config options.
const (
	DEFAULT_ADDR = "127.0.0.1:8900"
)

// ChatSrvStart is a goapp.AppStarter implementation which runs 
// a ChatSrv instance.
type ChatSrvStart struct {
	srv *chat.ChatSrv
}
func (this *ChatSrvStart) PreInit() {
	config.InitIniProvider("chat.ini", 1)
	debugLogs, _ := config.GetBoolVal("System.DebugLogs", 0, false)
	log.DebugLogs = debugLogs
}
func (this *ChatSrvStart) PostInit() {
	addr, _ := config.GetVal("ChatSrv.SrvAddr", 0, DEFAULT_ADDR)
	this.srv = chat.NewChatSrv(addr)
	this.srv.Start()
}

type ChatSrvLoop struct {}
func (this *ChatSrvLoop) OnHeartbeat() {
	netCounters := perf.GetCounterSet("Module.Net")
	msgRoute    := netCounters.Get(net.PERF_NET_MSG_ROUTE)

	chatCounters := perf.GetCounterSet("Module.Net.Proto.Chat")
	rcvRate      := chatCounters.Get(net.PERF_PROTO_RCV_OK)
	sendRate     := chatCounters.Get(net.PERF_PROTO_SEND_OK)

	perfStr := fmt.Sprintf(
		"route/sec: %d, rx/sec: %d, tx/sec: %d",
		msgRoute.PerSec(),
		rcvRate.PerSec(),
		sendRate.PerSec(),
	)

	log.Info(perfStr)
}
func (this *ChatSrvLoop) PreLoop() {}
func (this *ChatSrvLoop) PostLoop() {}


// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ChatSrvStart))
	goapp.SetLoopHandler(new(ChatSrvLoop))

	goapp.SetHeartbeat(1000) // 1000ms / 1sec

	stopChan := make(chan bool, 0)
	go goapp.Start("ChatSrv", stopChan)

	<-stopChan
}
