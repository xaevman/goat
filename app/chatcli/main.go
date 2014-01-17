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
	"github.com/xaevman/goat/lib/console"
	"github.com/xaevman/goat/lib/lifecycle"
	"github.com/xaevman/goat/proto/chat"
)

// Stdlib imports.
import (
	"fmt"
	"os"
	"strings"
)

// Console text constants.
const (
	EXIT_MSG   = "exit\n"
	PROMPT_TXT = "prompt> "
)

// Network objects.
var (
	msgProc 	= new(chat.MsgHandler)
	proto   	= net.NewProtocol("Chat")
	srv     	= net.NewTCPCli()
	srvAddr 	= "127.0.0.1:8900"
)

var discoChan          = net.NewDisconnectChan()
var currentChan uint32 = 0
var	myName             = "Anon"
var syncObj            = lifecycle.New()

// Text styles.
var (
	errStyle = console.Style {
		ForeColor: console.FG_RED,
		Bold:      true,
	}

	sysStyle = console.Style {
		ForeColor: console.FG_YELLOW,
		Bold:      true,
	}

	txtStyle = console.Style {
		ForeColor: console.FG_WHITE,
	}

	privStyle = console.Style {
		ForeColor: console.FG_MAGENTA,
	}
)


// main is the application entry point.
func main() {
	// args
    if len(os.Args) > 1 {
    	myName = os.Args[1]
    }

    // set up the console screen
	console.ClearScreen()

	// start network services
	proto.AddSignature(msgProc)
	proto.SetAccessProvider(new(net.NoSecurity))
	srv.RegisterDiscoHandler(discoChan)
	err := srv.Dial(srvAddr)
	if err != nil {
		console.WriteLineFmt(
			"Couldn't connect. Exiting...", 
			errStyle,
		)
		return
	}

	// get console input
	inChan := console.ReadInput(EXIT_MSG)

	// send connect packet
	sendConnect()

	// handle IO
	for syncObj.QueryRun() {
		select {
		case in := <-inChan:
			handleInput(in)
		case msg := <-msgProc.QueryReceiveMsg():
			handleMsg(msg)
		case <-discoChan.QueryDisconnect():
			onDisconnect()
		case <-syncObj.QueryShutdown():
		}
	}

	// shutdown
	proto.Shutdown()
	srv.Shutdown()

	// synchronize with background tasks shutting down
	syncObj.ShutdownComplete()
}

// handleInput validates console input, sends any local commands to the
// command handler, and passes off chat messages to the network layer for
// packing and transmission.
func handleInput(in string) {
	if in == EXIT_MSG {
		go syncObj.Shutdown()
		return
	}

	in = strings.TrimSpace(in)
	if len(in) < 1 {
		printPrompt(0)
		return
	}

	if in[0] == '/' {
		printText(
			fmt.Sprintf("Unsupported command %s", in),
			errStyle,
		)
		printPrompt(1)
		return
	}

	sendChat(currentChan, in)
}

// handleMsg reads incoming chat messages and redistributes them to the
// appropriate handlers.
func handleMsg(msg *chat.Msg) {
	log.Debug("%+v", msg)

	switch msg.Subtype {
	case chat.MSG_SUB_CHAT:
		printChatMsg(msg)
	case chat.MSG_SUB_CMD:
		return
	case chat.MSG_SUB_CONNECT:
		currentChan = msg.ChannelId
		sendSetName()
	case chat.MSG_SUB_JOIN_CHANNEL:
		return
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return
	case chat.MSG_SUB_SET_NAME:
		log.Debug(msg.Text)
	}
}

// onDisconnect is called when connection to the server is lost.
func onDisconnect() {
	go syncObj.Shutdown()

	console.WriteLine("")
	printText(
		"Disconnected from server", 
		errStyle,
	)
}

// printChatMsg prints an incoming chat message out on the client console.
func printChatMsg(msg *chat.Msg) {
	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)

	printText(
		fmt.Sprintf("\n<%s> %s", msg.From, msg.Text),
		txtStyle,
	)
	printPrompt(1)
}

// printPrompt inserts the given number of empty lines and then outputs the 
// prompt.
func printPrompt(spacing int) {
	for i := 0; i < spacing; i++ {
		console.WriteLine("")
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.SetBold()
	console.Write(PROMPT_TXT)
	console.ClearFormat()
}

// printText prints the given text, in the given style, on the line above
// the current cursor position.
func printText(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLineFmt(txt, style)
}

// sendChat sends a chat message to the server.
func sendChat(channel uint32, text string) {
	conMsg := chat.Msg {
		ChannelId: channel,
		From:      myName,
		Subtype:   chat.MSG_SUB_CHAT,
		ToId:      srv.Socket().Id(),
		Text:      text,
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}

// sendConnect sends the initial connection message to the server.
func sendConnect() {
	conMsg := chat.Msg {
		Subtype: chat.MSG_SUB_CONNECT,
		ToId:	 srv.Socket().Id(),
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}

// sendSetName sets this client's display name on the server.
func sendSetName() {
	conMsg := chat.Msg {
		From:    myName,
		Subtype: chat.MSG_SUB_SET_NAME,
		ToId:	 srv.Socket().Id(),
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}
