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
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/lifecycle"
	"github.com/xaevman/goat/proto/chat"
)

// Server objects.
var (
	msgProc = new(chat.MsgHandler)
	proto   = net.NewProtocol("Chat")
	srv     = net.NewTCPSrv()
	srvAddr = "127.0.0.1:8900"
	syncObj = lifecycle.New()
)

// main is the application entry point
func main() {
	log.DebugLogs = true

	// startup
	proto.AddSignature(msgProc)
	proto.SetAccessProvider(new(net.NoSecurity))
	srv.Start(srvAddr)

	// do stuff
	for syncObj.QueryRun() {
		select {
		case msg := <-msgProc.QueryReceiveMsg():
			handleMsg(msg)
		case <-syncObj.QueryShutdown():
		}
	}

	// shutdown
	proto.Shutdown()
	srv.Stop()

	syncObj.ShutdownComplete()
}

// handleMsg reads new chat messages and redistributes them to other clients
// in the same channel.
func handleMsg(msg *chat.Msg) {
	log.Info("%+v", msg)
}
