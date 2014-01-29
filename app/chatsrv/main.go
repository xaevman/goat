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
	"github.com/xaevman/goat/proto/chat"
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
	config.InitIniProvider("ChatSrv.ini", 1)	
}
func (this *ChatSrvStart) PostInit() {
	addr, _ := config.GetVal("ChatSrv.SrvAddr", 0, DEFAULT_ADDR)
	this.srv = chat.NewChatSrv(addr)
	this.srv.Start()
}

type ChatSrvLoop struct {}
func (this *ChatSrvLoop) OnHeartbeat() {
	netCounters := perf.GetCounterSet("Service.Net")
	msgRoute    := netCounters.Get(net.PERF_NET_MSG_ROUTE)

	chatCounters := perf.GetCounterSet("Service.Net.Proto.Chat")
	rcvRate      := chatCounters.Get(net.PERF_PROTO_RCV_OK)
	sendRate     := chatCounters.Get(net.PERF_PROTO_SEND_OK)

	log.Info("Route/Sec: %.2f", msgRoute.PerSec())
	log.Info("Rx/Sec: %.2f", rcvRate.PerSec())
	log.Info("Tx/Sec: %.2f", sendRate.PerSec())
}
func (this *ChatSrvLoop) PreLoop() {}
func (this *ChatSrvLoop) PostLoop() {}


// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ChatSrvStart))
	goapp.SetLoopHandler(new(ChatSrvLoop))

	goapp.SetHeartbeat(5000) // 5000ms / 5sec

	stopChan := make(chan bool, 0)
	go goapp.Start("ChatSrv", stopChan)

	<-stopChan
}
