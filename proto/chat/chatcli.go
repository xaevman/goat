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

package chat

// External imports.
import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/lifecycle"
)

// ChatCliNotifier defines the inteface for customized implementations
// of a chat client.
type ChatCliNotifier interface {
	OnMsg(msg *Msg)
	OnConnect(con net.Connection)
	OnDisconnect(con net.Connection)
	OnTimeout(timeout *net.TimeoutEvent)
	OnShutdown()
	SetParent(chatCli *ChatCli)
}

// ChatCli is the base client implementation for a chat client. This base
// object handles coordination
type ChatCli struct {
	addr        string
	evtHandler  *net.EventChan
	msgProc     *MsgHandler
	username    string
	notifier    ChatCliNotifier
	srv         *net.TCPCli
	syncObj     *lifecycle.Lifecycle
}

// NewChatCli is a helper constructor function which initializes a new
// ChatCli object and returns a pointer to it for use.
func NewChatCli(notify ChatCliNotifier) *ChatCli {
	cli := ChatCli {
		evtHandler : net.NewEventChan(),
		msgProc    : new(MsgHandler),
		username   : "Anon",
		notifier   : notify,
		srv        : net.NewTCPCli(),
		syncObj    : lifecycle.New(),
	}

	notify.SetParent(&cli)

	Protocol.AddSignature(cli.msgProc)
	Protocol.SetAccessProvider(new(net.NoSecurity))

	return &cli
}

// Connect registers the chat protocol and message objects, and then dials
// the specified server before starting the IO processing go routine.
func (this *ChatCli) Connect(addr string) error {
	go this.handleIO()

	return this.srv.Dial(addr)	
}

// Shutdown begins the shutdown process for the client and blocks until
// it completes.
func (this *ChatCli) Shutdown() {
	this.syncObj.Shutdown()
}

// SendChat sends a chat message to the server.
func (this *ChatCli) SendChat(channel uint32, text string) {
	conMsg          := new(Msg)
	conMsg.ChannelId = channel
	conMsg.From      = this.username
	conMsg.Subtype   = MSG_SUB_CHAT
	conMsg.ToId      = this.srv.Socket().Id()
	conMsg.Text      = text

	this.send(conMsg.ToId, conMsg)
}

// SendConnect sends the initial connection message to the server.
func (this *ChatCli) SendConnect() {
	conMsg        := new(Msg)
	conMsg.From    = this.username
	conMsg.Subtype = MSG_SUB_CONNECT
	conMsg.ToId    = this.srv.Socket().Id()

	this.send(conMsg.ToId, conMsg)
}

// SendJoinChannel sends a join channel request to the server.
func (this *ChatCli) SendJoinChannel(name string) {
	joinMsg        := new(Msg)
	joinMsg.Subtype = MSG_SUB_JOIN_CHANNEL
	joinMsg.Text    = name
	joinMsg.ToId    = this.srv.Socket().Id()

	this.send(joinMsg.ToId, joinMsg)
}

// SendSetName sets this client's display name on the server.
func (this *ChatCli) SendSetName() {
	conMsg        := new(Msg)
	conMsg.From    = this.username
	conMsg.Subtype = MSG_SUB_SET_NAME
	conMsg.ToId    = this.srv.Socket().Id()

	this.send(conMsg.ToId, conMsg)
}

// SetUsername sets the username for this client.
func (this *ChatCli) SetUsername(name string) {
	this.username = name
}

// handleIO runs in a separate go routine and handles all IO for the
// chat client.
func (this *ChatCli) handleIO() {
	for this.syncObj.QueryRun() {
		select {
		case msg := <-this.msgProc.QueryReceiveMsg():
			log.Debug("RX: %+v", msg)
			this.notifier.OnMsg(msg)
		case con := <-this.evtHandler.QueryConnect():
			this.notifier.OnConnect(con)
		case con := <-this.evtHandler.QueryDisconnect():
			this.notifier.OnDisconnect(con)
		case timeout := <-this.evtHandler.QueryTimeout():
			this.notifier.OnTimeout(timeout)
		case <-this.syncObj.QueryShutdown():
			this.notifier.OnShutdown()
		}
	}

	// shutdown
	Protocol.Shutdown()
	this.srv.Shutdown()

	// synchronize with background tasks shutting down
	this.syncObj.ShutdownComplete()
}

// send transmits a Msg object to the server.
func (this *ChatCli) send(toId uint32, msg *Msg) {
	log.Debug("TX: %+v", msg)
	this.msgProc.SendMsg(toId, msg)
}
