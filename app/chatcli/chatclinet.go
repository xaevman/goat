//  ---------------------------------------------------------------------------
//
//  chatclinet.go
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
	"github.com/xaevman/goat/prod"
	"github.com/xaevman/goat/prod/chat"
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


// ChatCli is an implementation of of net.EventHandler which implements
// a simple console-based chat client.
type ChatCli struct {
	chanIdMap   map[uint32]string
	chanNameMap map[string]uint32
	currentChan uint32
	inputSync   *lifecycle.Lifecycle
	proto       *net.Protocol
	srvId       uint32
	username    string
}

func (this *ChatCli) Close() {
	this.inputSync.Shutdown()
	console.Write(console.ENABLE_LINEWRAP)
	goapp.Stop()
}

func (this *ChatCli) Init(proto *net.Protocol) {
	this.chanIdMap   = make(map[uint32]string, 0)
	this.chanNameMap = make(map[string]uint32, 0)
	this.currentChan = 0
	this.inputSync   = lifecycle.New()
	this.proto       = proto

	this.proto.AddSignature(new(chat.MsgHandler))
	this.proto.SetAccessProvider(new(net.NoSecurity))

	go this.startInput()
}

// OnConnect is called when the connection is established with the server.
func (this *ChatCli) OnConnect(con net.Connection) {
	this.srvId = con.Id()

	console.ClearScreen()
	console.Write(console.DISABLE_LINEWRAP)
	console.WriteLine("====================================")
	console.WriteLine("ChatCli v0.1")
	console.WriteLine("Copyright 2014 Jared Chavez")
	console.WriteLine("====================================")
	console.WriteLine("")

	this.sendConnect()
}

// OnDisconnect is called when the connection to the server is dropped.
func (this *ChatCli) OnDisconnect(con net.Connection) {
	go func() {
		this.proto.Shutdown()

		console.WriteLine("")
		this.printChatText(
			"Disconnected from server",
			errStyle,
		)
	}()
}


func (this *ChatCli) OnError(err error) {
	log.Error(err.Error())
}

// OnMsg distributes chat messages received from the server to thier
// appropriate message handler.
func (this *ChatCli) OnReceive(msg interface{}) {
	chatMsg, ok := msg.(*chat.Msg)
	if !ok {
		log.Error("Invalid object type received (%T)", msg)
		return
	}

	switch chatMsg.Subtype {
	case chat.MSG_SUB_CHAT:
		this.printChatMsg(chatMsg)
	case chat.MSG_SUB_CMD:
		return
	case chat.MSG_SUB_CONNECT:
		this.onConnectMsg(chatMsg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		this.onJoinChannelMsg(chatMsg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return
	case chat.MSG_SUB_SET_NAME:
		this.printChatMsg(chatMsg)
	}
}

// OnShutdown is called when Shutdown() is called on the underlying
// ChatCli.
func (this *ChatCli) OnShutdown() {}

// OnTimeout is called when a TCP timeout event bubbles up from the
// net service.
func (this *ChatCli) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error("%+v", timeout)
}


// handleInput is called when a new line of input text is received from
// the console. If the supplied text matches EXIT_MSG, the shutdown
// sequence is started. Otherwise, the text is parsed as either a command
// (if starting with /) or a chat message and sent to the server
// appropriately.
func (this *ChatCli) handleInput(in string) {
	if in == EXIT_MSG {
		this.OnDisconnect(nil)
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

	this.sendChat(this.currentChan, in)
}

// onConnect is called on reciept of the connection confirmed packet from
// the server.
func (this *ChatCli) onConnectMsg(msg *chat.Msg) {
	con := this.proto.GetConnection(msg.FromId)
	txt := fmt.Sprintf("Connect (%v)", con.RemoteAddr())
	this.printChatText(txt, sysStyle)

	this.sendJoinChannel(chat.PUB_CHANNEL)
}

// onJoinChannel is called when a MSG_SUB_JOIN_CHANNEL message is received
// from the server.
func (this *ChatCli) onJoinChannelMsg(msg *chat.Msg) {
	this.chanIdMap[msg.ChannelId] = msg.From
	this.chanNameMap[msg.From]    = msg.ChannelId

	if msg.From == chat.PUB_CHANNEL {
		this.currentChan = msg.ChannelId
	}

	this.printPrompt(1)
}

// onLeaveChannel is called when a MSG_SUB_LEAVE_CHANNEL message is
// received from the server.
func (this *ChatCli) onLeaveChannel(msg *chat.Msg) {
	delete(this.chanIdMap,   msg.ChannelId)
	delete(this.chanNameMap, msg.From)
}

// printChatMsg prints an incoming chat message out on the client console.
func (this *ChatCli) printChatMsg(msg *chat.Msg) {
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

// printChatText prints any text which should go into the chat window.
// The lone exception is text input via Stdin, which should be handled by
// printTextFromInput because of the extra new line that will be input
// from that method.
func (this *ChatCli) printChatText(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")
	console.WriteLineFmt(txt, style)
}

// printPrompt inserts the given number of empty lines and then outputs the
// prompt.
func (this *ChatCli) printPrompt(spacing int) {
	for i := 0; i < spacing; i++ {
		console.WriteLine("")
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.SetBold()
	console.Write(this.chanIdMap[this.currentChan] + "> ")
	console.ClearFormat()
}

// printTextFromInput handles printing the given text after accounting
// for the extra new line that will be present from text input through
// Stdin.
func (this *ChatCli) printTextFromInput(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.WriteLineFmt(txt, style)
}

// send transmits a Msg object to the server.
func (this *ChatCli) send(msg *chat.Msg) {
	this.proto.SendMsg(this.srvId, products.CHAT_MSG, msg)
}

// SendChat sends a chat message to the server.
func (this *ChatCli) sendChat(channel uint32, text string) {
	conMsg          := new(chat.Msg)
	conMsg.ChannelId = channel
	conMsg.From      = this.username
	conMsg.Subtype   = chat.MSG_SUB_CHAT
	conMsg.Text      = text

	this.send(conMsg)
}

// SendConnect sends the initial connection message to the server.
func (this *ChatCli) sendConnect() {
	conMsg        := new(chat.Msg)
	conMsg.From    = this.username
	conMsg.Subtype = chat.MSG_SUB_CONNECT

	this.send(conMsg)
}

// SendJoinChannel sends a join channel request to the server.
func (this *ChatCli) sendJoinChannel(channelName string) {
	joinMsg        := new(chat.Msg)
	joinMsg.Subtype = chat.MSG_SUB_JOIN_CHANNEL
	joinMsg.Text    = channelName

	this.send(joinMsg)
}

// SendSetName sets this client's display name on the server.
func (this *ChatCli) sendSetName() {
	conMsg        := new(chat.Msg)
	conMsg.From    = this.username
	conMsg.Subtype = chat.MSG_SUB_SET_NAME

	this.send(conMsg)
}


func (this *ChatCli) startInput() {
	inChan := console.ReadInput(EXIT_MSG)

	for this.inputSync.QueryRun() {
		select {
		case in := <-inChan:
			this.handleInput(in)
		case <-this.inputSync.QueryShutdown():
		}
	}

	this.inputSync.ShutdownComplete()
}

