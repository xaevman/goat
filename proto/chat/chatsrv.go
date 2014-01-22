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

// Stdlib imports.
import (
	"fmt"
	"strings"
)

// ChatSrv implements the server side of of a chat application based on the
// protocol defined in proto/chat.
type ChatSrv struct {
	addr         string
	chanMap      map[uint32]*net.BroadcastGroup
	chanNameMap  map[string]*net.BroadcastGroup
	evtHandler   *net.EventChan
	msgProc      *MsgHandler
	shutdownChan chan bool
	srv          *net.TCPSrv
	syncObj      *lifecycle.Lifecycle
	userMap      map[uint32]string
}

// NewChatSrv initializes a new copy of ChatSrv and returns a pointer to it
// for use.
func NewChatSrv(addr string) *ChatSrv {
	chatSrv := ChatSrv {
		addr         : addr,
		chanMap      : make(map[uint32]*net.BroadcastGroup, 0),
		chanNameMap  : make(map[string]*net.BroadcastGroup, 0),
		evtHandler   : net.NewEventChan(),
		msgProc      : new(MsgHandler),
		shutdownChan : make(chan bool, 0),
		srv          : net.NewTCPSrv(),
	    syncObj      : lifecycle.New(),
		userMap      : make(map[uint32]string, 0),
	}

	Protocol.AddSignature(chatSrv.msgProc)
	Protocol.SetAccessProvider(new(net.NoSecurity))

	return &chatSrv
}

// Start initializes and starts the ChatSrv instance.
func (this *ChatSrv) Start() {
	go this.handleIO()
	go this.srv.Start(this.addr)
}

// Stop begins the shutdown process of the server, blocking until complete.
func (this *ChatSrv) Stop() {
	this.syncObj.Shutdown()
}

// announceUser sends a message announcing a user's presence to all other users
// in a channel.
func (this *ChatSrv) announceUserJoin(ch *net.BroadcastGroup, user uint32) {
	msg        := new(Msg)
	msg.Subtype = MSG_SUB_CHAT
	msg.ToId    = ch.Id()
	msg.Text    = fmt.Sprintf(
		"%s joined channel '%s'",
		this.userMap[user],
		ch.Name(),
	)

	this.send(msg.ToId, msg)
}

// announceUserLeave sends a message announcing a user's departure to all other
// users in a channel.
func (this *ChatSrv) announceUserLeave(ch *net.BroadcastGroup, user uint32) {
	msg        := new(Msg)
	msg.Subtype = MSG_SUB_CHAT
	msg.ToId    = ch.Id()
	msg.Text    = fmt.Sprintf(
		"%s left channel '%s'",
		this.userMap[user],
		ch.Name(),
	)

	this.send(msg.ToId, msg)
}

// createChannel checks the registration maps for the named channel and either
// returns a pointer to an existing channel, or creates, registers and returns
// a pointer to a new channel object.
func (this *ChatSrv) createChannel(name string) *net.BroadcastGroup {
	ch := this.chanNameMap[name]
	if ch == nil {
		log.Debug("Creating channel %s", name)
		ch = net.NewBroadcastGroup(name)
		this.chanMap[ch.Id()] = ch
		this.chanNameMap[ch.Name()] = ch

		Protocol.RegisterConnection(ch)
	}

	return ch
}

// distChatMsg distributes an incoming chat message to all clients in the given
// channel.
func (this *ChatSrv) distChatMsg(msg *Msg) {
	ch := this.chanMap[msg.ChannelId]
	if ch == nil {
		// channel doesn't exist
		return
	}

	if ch.GetConnection(msg.FromId) == nil {
		// user not a member of that channel
		return
	}

	this.send(msg.ChannelId, msg)
}

// handleDisco removes disconnected clients from channel lists.
func (this *ChatSrv) handleDisco(con net.Connection) {
	for k, _ := range this.chanMap {
		ch    := this.chanMap[k]
		chCon := ch.GetConnection(con.Id())
		if chCon == con {
			ch.RemoveConnection(con.Id())
			this.announceUserLeave(ch, con.Id())
		}
	}

	delete(this.userMap, con.Id())
}

// handleIO runs in a separate go routine and handles all network IO
// and events from the network service.
func (this *ChatSrv) handleIO() {
	// application loop
	for this.syncObj.QueryRun() {
		select {
		case msg := <-this.msgProc.QueryReceiveMsg():
			this.handleMsg(msg)
		case <-this.evtHandler.QueryConnect():
			continue
		case con := <-this.evtHandler.QueryDisconnect():
			this.handleDisco(con)
		case timeout := <-this.evtHandler.QueryTimeout():
			log.Error("%+v", timeout)
		case <-this.syncObj.QueryShutdown():
		}
	}

	// shutdown signaled
	Protocol.Shutdown()
	this.srv.Stop()

	// unblock this.Stop()
	this.syncObj.ShutdownComplete()
}

// handleMsg reads new messages and redistributes them to their appropriate
// handlers based on subtype.
func (this *ChatSrv) handleMsg(msg *Msg) {
	log.Debug("RX: %+v", msg)

	switch msg.Subtype {
	case MSG_SUB_CHAT:
		this.distChatMsg(msg)
	case MSG_SUB_CMD:
		return
	case MSG_SUB_CONNECT:
		this.handleConnect(msg)
	case MSG_SUB_JOIN_CHANNEL:
		this.handleJoinChan(msg)
	case MSG_SUB_LEAVE_CHANNEL:
		return
	case MSG_SUB_SET_NAME:
		this.handleSetName(msg)
	}
}

// handleConnect sends a message back to the conneting client to confirm
// their access and set their initial channel.
func (this *ChatSrv) handleConnect(msg *Msg) {
	toId              := msg.FromId
	this.userMap[toId] = msg.From

	// reply
	msg.ChannelId = 0
	msg.From      = ""
	msg.ToId      = toId

	log.Debug("TX: %+v", msg)
	go this.msgProc.SendMsg(toId, msg)

	// send welcome message
	this.sendWelcomeMsg(toId, this.userMap[toId])
}

// handleJoinChan adds a user to a channel, if not already presents and they
// have required rights, and then calls accounceUser().
func (this *ChatSrv) handleJoinChan(msg *Msg) {
	ch := this.createChannel(strings.TrimSpace(msg.Text))
	userId := msg.FromId

	con := ch.GetConnection(userId)
	if con == nil {
		ch.AddConnection(Protocol.GetConnection(userId))
		this.announceUserJoin(ch, userId)
	}

	// reply
	msg.ChannelId = ch.Id()
	msg.From      = ch.Name()
	msg.ToId      = userId

	this.send(msg.ToId, msg)
}

// handleSetName links a client network ID to a friendly username. If the client
// is registering for the first time, a welcome message is sent back to that
// client.
func (this *ChatSrv) handleSetName(msg *Msg) {
	this.userMap[msg.FromId] = msg.From
	this.sendSetNameConfirm(msg)
}

// send transmits a new message through the message processor.
func (this *ChatSrv) send(toId uint32, msg *Msg) {
	log.Debug("TX: %+v", msg)
	this.msgProc.SendMsg(toId, msg)
}

// sendJoinChanConfirm sends a message to the client to let it know that it has
// successfully joined a new channel.
func (this *ChatSrv) sendJoinChanConfirm(ch *net.BroadcastGroup, toId uint32) {
	msg          := new(Msg)
	msg.ChannelId = ch.Id()
	msg.From      = ch.Name()
	msg.Subtype   = MSG_SUB_JOIN_CHANNEL
	msg.ToId      = toId

	this.send(toId, msg)
}

// sendSetNameConfirm sends a confirmation that the user's name was set
// successfully.
func (this *ChatSrv) sendSetNameConfirm(msg *Msg) {
	msg.Text = fmt.Sprintf("Name set to %v", msg.From)
	msg.From = ""
	msg.ToId = msg.FromId

	this.send(msg.ToId, msg)
}

// sendWelcomeMsg sends a quick hello to newly connected users.
func (this *ChatSrv) sendWelcomeMsg(id uint32, name string) {
	conMsg          := new(Msg)
	conMsg.ChannelId = 0
	conMsg.Subtype   = MSG_SUB_CHAT
	conMsg.ToId      = id
	conMsg.Text      = fmt.Sprintf("Welcome %v!", name)

	this.send(conMsg.ToId, conMsg)
}
