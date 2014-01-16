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

// Stdlib imports.
import (
	"bufio"
	"os"
	"strings"
)

// Server objects.
var (
	consoleChan = make(chan string)
	msgProc 	= new(chat.MsgHandler)
	myName  	= "Anon"
	proto   	= net.NewProtocol("Chat")
	srv     	= net.NewTCPCli()
	srvAddr 	= "127.0.0.1:8900"
	syncObj 	= lifecycle.New()
)

// main is the application entry point
func main() {
	log.DebugLogs = true

	// args
    if len(os.Args) > 1 {
    	myName = os.Args[1]
    }

    go runConsole()

	// net startup
	proto.AddSignature(msgProc)
	proto.SetAccessProvider(new(net.NoSecurity))
	err := srv.Dial(srvAddr)
	if err != nil {
		panic(err)
	}

	sendConnect()

	// do stuff
	for syncObj.QueryRun() {
		select {
		case in := <-consoleChan:
			handleInput(in)
		case msg := <-msgProc.QueryReceiveMsg():
			handleMsg(msg)
		case <-syncObj.QueryShutdown():
		}
	}

	// shutdown
	proto.Shutdown()
	srv.Shutdown()

	syncObj.ShutdownComplete()
}

// handleInput translates console text to chat.Msg objects and sends them to
// the server.
func handleInput(in string) {
	if in[0] == '/' {
		log.Debug("Unsupported cmd: %s", strings.TrimSpace(in[1:]))
		return
	}

	conMsg := chat.Msg {
		From:    myName,
		Subtype: chat.MSG_SUB_CHAT,
		ToId:    srv.Socket().Id(),
		Text:    strings.TrimSpace(in),
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}

// handleMsg reads new chat messages and redistributes them to other clients
// in the same channel.
func handleMsg(msg *chat.Msg) {
	if msg.Subtype != chat.MSG_SUB_CHAT {
		return
	}

	log.Info("%+v", msg)
}

// runConsole scans Stdin for lines of text and sends them to the command
// processor
func runConsole() {
    console := bufio.NewReader(os.Stdin)
    for syncObj.QueryRun() {
    	txt, err := console.ReadString('\n')
    	if err != nil {
    		panic(err)
    	}

    	consoleChan<- txt
	}
}

// sendConnect sends the initial connection message to the server
func sendConnect() {
	conMsg := chat.Msg {
		From:    myName,
		Subtype: chat.MSG_SUB_CONNECT,
		ToId:	 srv.Socket().Id(),
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}
