//  ---------------------------------------------------------------------------
//
//  goapp.go
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
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/console"
	"github.com/xaevman/goat/lib/lifecycle"
	"github.com/xaevman/goat/proto/chat"
)

// Stdlib imports.
import (
	"fmt"
	"strings"
)

// Console text constants.
const (
	EXIT_MSG = "exit\n"
)

// Text styles.
var (
	errStyle = console.Style{
		ForeColor: console.FG_RED,
		Bold:      true,
	}

	sysStyle = console.Style{
		ForeColor: console.FG_YELLOW,
		Bold:      true,
	}

	txtStyle = console.Style{
		ForeColor: console.FG_WHITE,
	}

	privStyle = console.Style{
		ForeColor: console.FG_MAGENTA,
	}
)

// Object maps.
var (
	chanIdMap   = make(map[uint32]string, 0)
	chanNameMap = make(map[string]uint32, 0)
)


// ConsoleCli is an implementation of of ChatCliNotifier which implements
// a simple console-based chat client.
type ConsoleCli struct {
	chanIdMap   map[uint32]string
	chanNameMap map[string]uint32
	currentChan uint32
	inputSync   *lifecycle.Lifecycle
	parent      *chat.ChatCli
}

// NewConsoleCli is a constructor helper which initializes a new instance
// of ConsoleCli and returns a pointer to it for use.
func NewConsoleCli() *ConsoleCli {
	ccli := ConsoleCli {
		chanIdMap   : make(map[uint32]string, 0),
		chanNameMap : make(map[string]uint32, 0),
		currentChan : 0,
		inputSync   : lifecycle.New(),
	}

	go ccli.startInput()

	return &ccli
}

// OnMsg distributes chat messages received from the server to thier
// appropriate message handler.
func (this *ConsoleCli) OnMsg(msg *chat.Msg) {
	switch msg.Subtype {
	case chat.MSG_SUB_CHAT:
		this.printChatMsg(msg)
	case chat.MSG_SUB_CMD:
		return
	case chat.MSG_SUB_CONNECT:
		this.onConnectMsg(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		this.onJoinChannelMsg(msg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return
	case chat.MSG_SUB_SET_NAME:
		this.printChatMsg(msg)
	}
}

// OnConnect is called when the connection is established with the server.
func (this *ConsoleCli) OnConnect(con net.Connection) {
	console.ClearScreen()
	console.Write(console.DISABLE_LINEWRAP)
	console.WriteLine("====================================")
	console.WriteLine("ChatCli v0.1")
	console.WriteLine("Copyright 2014 Jared Chavez")
	console.WriteLine("====================================")
	console.WriteLine("")

	this.parent.SendConnect()
}

// OnDisconnect is called when the connection to the server is dropped.
func (this *ConsoleCli) OnDisconnect(con net.Connection) {
	go this.parent.Shutdown()

	console.WriteLine("")
	this.printChatText(
		"Disconnected from server",
		errStyle,
	)

	goapp.Stop()
}

// OnTimeout is called when a TCP timeout event bubbles up from the
// net service.
func (this *ConsoleCli) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error("%+v", timeout)
}

// OnShutdown is called when Shutdown() is called on the underlying
// ChatCli.
func (this *ConsoleCli) OnShutdown() {
	console.Write(console.ENABLE_LINEWRAP)
	goapp.Stop()
}

// SetParent is called by a ChatCli object when this object is
// registered with it, in order to give an opportunity to save a reference
// to the parent object.
func (this *ConsoleCli) SetParent(chatCli *chat.ChatCli) {
	this.parent = chatCli
}

// handleInput is called when a new line of input text is received from
// the console. If the supplied text matches EXIT_MSG, the shutdown
// sequence is started. Otherwise, the text is parsed as either a command
// (if starting with /) or a chat message and sent to the server
// appropriately.
func (this *ConsoleCli) handleInput(in string) {
	if in == EXIT_MSG {
		this.parent.Shutdown()
		return
	}

	in = strings.TrimSpace(in)
	if len(in) < 1 {
		this.printPrompt(0)
		return
	}

	if in[0] == '/' {
		this.printTextFromInput(
			fmt.Sprintf("Unsupported command %s", in),
			errStyle,
		)
		this.printPrompt(1)
		return
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	this.printPrompt(1)

	this.parent.SendChat(this.currentChan, in)
}

// onConnect is called on reciept of the connection confirmed packet from
// the server.
func (this *ConsoleCli) onConnectMsg(msg *chat.Msg) {
	con := chat.Protocol.GetConnection(msg.FromId)

	txt := fmt.Sprintf("Connect (%v)", con.RemoteAddr())
	this.printChatText(txt, sysStyle)

	this.parent.SendJoinChannel(chat.PUB_CHANNEL)
}

// onJoinChannel is called when a MSG_SUB_JOIN_CHANNEL message is received
// from the server.
func (this *ConsoleCli) onJoinChannelMsg(msg *chat.Msg) {
	this.chanIdMap[msg.ChannelId] = msg.From
	this.chanNameMap[msg.From]    = msg.ChannelId

	if msg.From == chat.PUB_CHANNEL {
		this.currentChan = msg.ChannelId
	}

	this.printPrompt(1)
}

// onLeaveChannel is called when a MSG_SUB_LEAVE_CHANNEL message is
// received from the server.
func (this *ConsoleCli) onLeaveChannel(msg *chat.Msg) {
	delete(this.chanIdMap,   msg.ChannelId)
	delete(this.chanNameMap, msg.From)
}

// printChatMsg prints an incoming chat message out on the client console.
func (this *ConsoleCli) printChatMsg(msg *chat.Msg) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")

	switch {
	case msg.ChannelId == 0:
		this.printChatText(msg.Text, sysStyle)
	case msg.ChannelId != this.currentChan:
		this.printChatText(
			fmt.Sprintf(
				"<%s.%s> %s",
				this.chanIdMap[msg.ChannelId],
				msg.From,
				msg.Text,
			),
			privStyle,
		)
	default:
		this.printChatText(
			fmt.Sprintf(
				"<%s.%s> %s",
				this.chanIdMap[msg.ChannelId],
				msg.From,
				msg.Text,
			),
			txtStyle,
		)
	}

	this.printPrompt(1)
}

// printPrompt inserts the given number of empty lines and then outputs the
// prompt.
func (this *ConsoleCli) printPrompt(spacing int) {
	for i := 0; i < spacing; i++ {
		console.WriteLine("")
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.SetBold()
	console.Write(this.chanIdMap[this.currentChan] + "> ")
	console.ClearFormat()
}

// printChatText prints any text which should go into the chat window.
// The lone exception is text input via Stdin, which should be handled by
// printTextFromInput because of the extra new line that will be input
// from that method.
func (this *ConsoleCli) printChatText(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")
	console.WriteLineFmt(txt, style)
}

// printTextFromInput handles printing the given text after accounting
// for the extra new line that will be present from text input through
// Stdin.
func (this *ConsoleCli) printTextFromInput(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.WriteLineFmt(txt, style)
}

func (this *ConsoleCli) startInput() {
	inChan := console.ReadInput(EXIT_MSG)

	for this.inputSync.QueryRun() {
		select {
		case in := <-inChan:
			this.handleInput(in)
		case <-this.inputSync.QueryShutdown():
		}
	}
}
