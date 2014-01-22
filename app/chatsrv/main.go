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

// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ChatSrvStart))

	stopChan := make(chan bool, 0)
	go goapp.Start("ChatSrv", stopChan)

	<-stopChan
}
