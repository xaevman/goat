//  ---------------------------------------------------------------------------
//
//  chatsrvnet.go
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
import(
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/prod"
	"github.com/xaevman/goat/prod/chat"
)

// Stdlib imports.
import(
	"fmt"
	"strings"
)


// ChatSrv is a net.EventHandler implementation which implements the
// behaviors of a basic chat server.
type ChatSrv struct {
	chanMap     map[uint32]*net.BroadcastGroup
	chanNameMap map[string]*net.BroadcastGroup
	proto       *net.Protocol
	userMap     map[uint32]string
}

// Close perofrms no operations in ChatSrv.
func (this *ChatSrv) Close() {}

// Init creates the internal maps which track chat channels and user, and
// also sets up chat message handler and server security handler.
func (this *ChatSrv) Init(proto *net.Protocol) {
	this.chanMap     = make(map[uint32]*net.BroadcastGroup, 0)
	this.chanNameMap = make(map[string]*net.BroadcastGroup, 0)
	this.proto       = proto
	this.userMap     = make(map[uint32]string)

	proto.AddSignature(new(chat.MsgHandler))
	proto.SetAccessProvider(new(net.NoSecurity))
}

// OnConnect logs debugging information about a newly connected client.
func (this *ChatSrv) OnConnect(con net.Connection) {
	log.Debug("OnConnect event: %s", con.RemoteAddr())
}

// OnDisconnect logs debugging information about a newly disconnected client
// and also removes the client from relevant channels and maps.
func (this *ChatSrv) OnDisconnect(con net.Connection) {
	log.Debug("OnDisconnect event: %s", con.RemoteAddr())
	this.handleDisco(con)
}

// OnError passes any errors from the network layer on to the logging system.
func (this *ChatSrv) OnError(err error) {
	log.Error(err.Error())
}

// OnReceive logs debugging information about incoming messages, performs
// a type assertion on the incoming message object, and then passes it to
// the message handler.
func (this *ChatSrv) OnReceive(msg interface{}) {
	log.Debug("%v", msg)

	chatMsg, ok := msg.(*chat.Msg)
	if !ok {
		log.Error("Invalid type received: %T", msg)
		return
	}

	this.handleMsg(chatMsg)
}

// OnShutdown logs the shutdown event.
func (this *ChatSrv) OnShutdown() {
	log.Info("Shutdown signal received")
}

// OnTimeout logs the occurance of timeout events. It makes no attempts to
// retry.
func (this *ChatSrv) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error(
		"Timeout [%d], msgType: %d", 
		timeout.TimeoutType,
		timeout.MessageType,
	)
}


// announceUser sends a message announcing a user's presence to all other users
// in a channel.
func (this *ChatSrv) announceUserJoin(ch *net.BroadcastGroup, user uint32) {
	msg        := new(chat.Msg)
	msg.Subtype = chat.MSG_SUB_CHAT
	msg.Text    = fmt.Sprintf(
		"%s joined channel '%s'",
		this.userMap[user],
		ch.Name(),
	)

	this.send(ch.Id(), msg)
}

// announceUserLeave sends a message announcing a user's departure to all other
// users in a channel.
func (this *ChatSrv) announceUserLeave(ch *net.BroadcastGroup, user uint32) {
	msg        := new(chat.Msg)
	msg.Subtype = chat.MSG_SUB_CHAT
	msg.Text    = fmt.Sprintf(
		"%s left channel '%s'",
		this.userMap[user],
		ch.Name(),
	)

	this.send(ch.Id(), msg)
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

		this.proto.RegisterConnection(ch)

		perfs.Increment(PERF_CHATSRV_CREATE_CHAN)
	}

	return ch
}

// distChatMsg distributes an incoming chat message to all clients in the given
// channel.
func (this *ChatSrv) distChatMsg(msg *chat.Msg) {
	ch := this.chanMap[msg.ChannelId]
	if ch == nil {
		// channel doesn't exist
		return
	}

	if ch.GetConnection(msg.FromId) == nil {
		// user not a member of that channel
		return
	}

	msg.From = this.userMap[msg.FromId]

	this.send(msg.ChannelId, msg)
}

// handleConnect sends a message back to the conneting client to confirm
// their access and set their initial channel.
func (this *ChatSrv) handleConnect(msg *chat.Msg) {
	this.userMap[msg.FromId] = msg.From

	// reply
	msg.ChannelId = 0
	msg.From      = ""

	this.send(msg.FromId, msg)

	// send welcome message
	this.sendWelcomeMsg(msg.FromId, this.userMap[msg.FromId])
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

// handleJoinChan adds a user to a channel, if not already presents and they
// have required rights, and then calls accounceUser().
func (this *ChatSrv) handleJoinChan(msg *chat.Msg) {
	ch := this.createChannel(strings.TrimSpace(msg.Text))
	userId := msg.FromId

	con := ch.GetConnection(userId)
	if con == nil {
		ch.AddConnection(this.proto.GetConnection(userId))
		this.announceUserJoin(ch, userId)
	}

	// reply
	msg.ChannelId = ch.Id()
	msg.From      = ch.Name()

	this.send(userId, msg)
}

// handleMsg reads new messages and redistributes them to their appropriate
// handlers based on subtype.
func (this *ChatSrv) handleMsg(msg *chat.Msg) {
	switch msg.Subtype {
	case chat.MSG_SUB_CHAT:
		perfs.Increment(PERF_CHATSRV_DIST_MSG)
		this.distChatMsg(msg)
	case chat.MSG_SUB_CMD:
		return
	case chat.MSG_SUB_CONNECT:
		perfs.Increment(PERF_CHATSRV_CONNECT_MSG)
		this.handleConnect(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		perfs.Increment(PERF_CHATSRV_JOIN_CHAN)
		this.handleJoinChan(msg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		perfs.Increment(PERF_CHATSRV_LEAVE_CHAN)
		return
	case chat.MSG_SUB_SET_NAME:
		perfs.Increment(PERF_CHATSRV_SET_NAME)
		this.handleSetName(msg)
	}
}

// handleSetName links a client network ID to a friendly username. If the client
// is registering for the first time, a welcome message is sent back to that
// client.
func (this *ChatSrv) handleSetName(msg *chat.Msg) {
	this.userMap[msg.FromId] = msg.From
	this.sendSetNameConfirm(msg)
}

// send transmits a new message through the message processor.
func (this *ChatSrv) send(toId uint32, msg *chat.Msg) {
	this.proto.SendMsg(toId, prod.CHAT_MSG, msg)
}

// sendJoinChanConfirm sends a message to the client to let it know that it has
// successfully joined a new channel.
func (this *ChatSrv) sendJoinChanConfirm(ch *net.BroadcastGroup, toId uint32) {
	msg          := new(chat.Msg)
	msg.ChannelId = ch.Id()
	msg.From      = ch.Name()
	msg.Subtype   = chat.MSG_SUB_JOIN_CHANNEL

	this.send(toId, msg)
}

// sendSetNameConfirm sends a confirmation that the user's name was set
// successfully.
func (this *ChatSrv) sendSetNameConfirm(msg *chat.Msg) {
	msg.Text = fmt.Sprintf("Name set to %v", msg.From)
	msg.From = ""

	this.send(msg.FromId, msg)
}

// sendWelcomeMsg sends a quick hello to newly connected users.
func (this *ChatSrv) sendWelcomeMsg(id uint32, name string) {
	conMsg          := new(chat.Msg)
	conMsg.ChannelId = 0
	conMsg.Subtype   = chat.MSG_SUB_CHAT
	conMsg.Text      = fmt.Sprintf("Welcome %v!", name)

	this.send(id, conMsg)
}
