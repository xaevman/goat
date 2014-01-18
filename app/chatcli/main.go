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
	msgProc = new(chat.MsgHandler)
	proto   = net.NewProtocol("Chat")
	srv     = net.NewTCPCli()
	srvAddr = "127.0.0.1:8900"
)

// Object maps.
var (
	chanIdMap   = make(map[uint32]string, 0)
	chanNameMap = make(map[string]uint32, 0)
)

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

var discoChan          = net.NewDisconnectChan()
var	myName             = "Anon"
var currentChan uint32 = 0
var syncObj            = lifecycle.New()


// main is the application entry point.
func main() {
	//log.DebugLogs = true

	// args
	if len(os.Args) > 1 {
		srvAddr = os.Args[1]
	}
    if len(os.Args) > 2 {
    	myName = os.Args[2]
    }

    // set up the console screen
	console.ClearScreen()
	console.WriteLine("====================================")
	console.WriteLine("ChatCli v0.1")
	console.WriteLine("Copyright 2014 Jared Chavez")
	console.WriteLine("====================================")
	console.WriteLine("")

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
		printTextFromInput(
			fmt.Sprintf("Unsupported command %s", in),
			errStyle,
		)
		printPrompt(1)
		return
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	printPrompt(1)

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
		onConnect(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		onJoinChannel(msg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		onLeaveChannel(msg)
	case chat.MSG_SUB_SET_NAME:
		printChatMsg(msg)
	}
}

// onConnect is called on reciept of teh connection confirmed packet from
// the server.
func onConnect(msg *chat.Msg) {
	con := proto.GetConnection(msg.FromId)

	txt := fmt.Sprintf("Connect (%v)", con.RemoteAddr())
	printChatText(txt, sysStyle)

	sendJoinChannel(chat.PUB_CHANNEL)
}

// onDisconnect is called when connection to the server is lost.
func onDisconnect() {
	go syncObj.Shutdown()

	console.WriteLine("")
	printChatText(
		"Disconnected from server", 
		errStyle,
	)
}

// onJoinChannel is called when a MSG_SUB_JOIN_CHANNEL message is received
// from the server.
func onJoinChannel(msg *chat.Msg) {
	chanIdMap[msg.ChannelId] = msg.From
	chanNameMap[msg.From]    = msg.ChannelId

	if msg.From == chat.PUB_CHANNEL {
		currentChan = msg.ChannelId
	}

	printPrompt(1)
}

// onLeaveChannel is called when a MSG_SUB_LEAVE_CHANNEL message is 
// received from the server.
func onLeaveChannel(msg *chat.Msg) {
	delete(chanIdMap, msg.ChannelId)
	delete(chanNameMap, msg.From)
}

// printChatMsg prints an incoming chat message out on the client console.
func printChatMsg(msg *chat.Msg) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")

	switch {
	case msg.ChannelId == 0:
		printChatText(msg.Text, sysStyle)
	case msg.ChannelId != currentChan:
		printChatText(
			fmt.Sprintf(
				"<%s.%s> %s", 
				chanIdMap[msg.ChannelId],
				msg.From, 
				msg.Text,
			),
			privStyle,
		)
	default:
		printChatText(
			fmt.Sprintf(
				"<%s.%s> %s", 
				chanIdMap[msg.ChannelId], 
				msg.From, 
				msg.Text,
			),
			txtStyle,
		)
	}

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
	console.Write(chanIdMap[currentChan] + "> ")
	console.ClearFormat()
}

// printChatText prints any text which should go into the chat window.
// The lone exception is text input via Stdin, which should be handled by
// printTextFromInput because of the extra new line that will be input
// from that method.
func printChatText(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")
	console.WriteLineFmt(txt, style)
}

// printTextFromInput handles printing the given text after accounting
// for the extra new line that will be present from text input through
// Stdin.
func printTextFromInput(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.WriteLineFmt(txt, style)
}

// sendChat sends a chat message to the server.
func sendChat(channel uint32, text string) {
	conMsg          := new(chat.Msg)
	conMsg.ChannelId = channel
	conMsg.From      = myName
	conMsg.Subtype   = chat.MSG_SUB_CHAT
	conMsg.ToId      = srv.Socket().Id()
	conMsg.Text      = text

	msgProc.SendMsg(conMsg.ToId, conMsg)
}

// sendConnect sends the initial connection message to the server.
func sendConnect() {
	conMsg        := new(chat.Msg)
	conMsg.From    = myName
	conMsg.Subtype = chat.MSG_SUB_CONNECT
	conMsg.ToId    = srv.Socket().Id()

	msgProc.SendMsg(conMsg.ToId, conMsg)
}

// sendJoinChannel sends a join channel request to the server.
func sendJoinChannel(name string) {
	joinMsg        := new(chat.Msg)
	joinMsg.Subtype = chat.MSG_SUB_JOIN_CHANNEL
	joinMsg.Text    = name
	joinMsg.ToId    = srv.Socket().Id()

	msgProc.SendMsg(joinMsg.ToId, joinMsg)
}

// sendSetName sets this client's display name on the server.
func sendSetName() {
	conMsg        := new(chat.Msg)
	conMsg.From    = myName
	conMsg.Subtype = chat.MSG_SUB_SET_NAME
	conMsg.ToId    = srv.Socket().Id()

	msgProc.SendMsg(conMsg.ToId, conMsg)
}
