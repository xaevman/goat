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
	"fmt"
	"os"
	"strings"
)

// Network objects.
var (
	evtHandler = net.NewEventChan()
	msgProc    = new(chat.MsgHandler)
	proto      = net.NewProtocol("Chat")
	srv        = net.NewTCPSrv()
	srvAddr    = "127.0.0.1:8900"
)

// Synchronization helpers.
var syncObj = lifecycle.New()

// Object maps.
var (
	chanMap     = make(map[uint32]*net.BroadcastGroup, 0)
	chanNameMap = make(map[string]*net.BroadcastGroup, 0)
	userMap     = make(map[uint32]string, 0)
)

// main is the application entry point
func main() {
	log.DebugLogs = true

	// args
	if len(os.Args) > 1 {
		srvAddr = os.Args[1]
	}

	// startup
	proto.AddSignature(msgProc)
	proto.SetAccessProvider(new(net.NoSecurity))
	srv.Start(srvAddr)

	// do stuff
	for syncObj.QueryRun() {
		select {
		case msg := <-msgProc.QueryReceiveMsg():
			handleMsg(msg)
		case <-evtHandler.QueryConnect():
			continue
		case con := <-evtHandler.QueryDisconnect():
			handleDisco(con)
		case timeout := <-evtHandler.QueryTimeout():
			log.Error("%+v", timeout)
		case <-syncObj.QueryShutdown():
		}
	}

	// shutdown
	proto.Shutdown()
	srv.Stop()

	syncObj.ShutdownComplete()
}

// announceUser sends a message announcing a user's presence to all other users
// in a channel.
func announceUserJoin(ch *net.BroadcastGroup, user uint32) {
	msg := new(chat.Msg)
	msg.Subtype = chat.MSG_SUB_CHAT
	msg.ToId = ch.Id()
	msg.Text = fmt.Sprintf(
		"%s joined channel '%s'",
		userMap[user],
		ch.Name(),
	)

	log.Debug("%+v", msg)
	msgProc.SendMsg(msg.ToId, msg)
}

// announceUserLeave sends a message announcing a user's departure to all other
// users in a channel.
func announceUserLeave(ch *net.BroadcastGroup, user uint32) {
	msg := new(chat.Msg)
	msg.Subtype = chat.MSG_SUB_CHAT
	msg.ToId = ch.Id()
	msg.Text = fmt.Sprintf(
		"%s left channel '%s'",
		userMap[user],
		ch.Name(),
	)

	log.Debug("%+v", msg)
	msgProc.SendMsg(msg.ToId, msg)
}

// createChannel checks the registration maps for the named channel and either
// returns a pointer to an existing channel, or creates, registers and returns
// a pointer to a new channel object.
func createChannel(name string) *net.BroadcastGroup {
	ch := chanNameMap[name]
	if ch == nil {
		log.Debug("Creating channel %s", name)
		ch = net.NewBroadcastGroup(name)
		chanMap[ch.Id()] = ch
		chanNameMap[ch.Name()] = ch

		proto.RegisterConnection(ch)
	}

	return ch
}

// distChatMsg distributes an incoming chat message to all clients in the given
// channel.
func distChatMsg(msg *chat.Msg) {
	log.Debug("%+v", msg)

	ch := chanMap[msg.ChannelId]
	if ch == nil {
		// channel doesn't exist
		return
	}

	if ch.GetConnection(msg.FromId) == nil {
		// user not a member of that channel
		return
	}

	msgProc.SendMsg(msg.ChannelId, msg)
}

// handleDisco removes disconnected clients from channel lists.
func handleDisco(con net.Connection) {
	for k, _ := range chanMap {
		ch := chanMap[k]
		chCon := ch.GetConnection(con.Id())
		if chCon == con {
			ch.RemoveConnection(con.Id())
			announceUserLeave(ch, con.Id())
		}
	}

	delete(userMap, con.Id())
}

// handleMsg reads new chat messages and redistributes them to other clients
// in the same channel.
func handleMsg(msg *chat.Msg) {
	log.Debug("%+v", msg)

	switch msg.Subtype {
	case chat.MSG_SUB_CHAT:
		distChatMsg(msg)
	case chat.MSG_SUB_CMD:
		return
	case chat.MSG_SUB_CONNECT:
		handleConnect(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		handleJoinChan(msg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return
	case chat.MSG_SUB_SET_NAME:
		handleSetName(msg)
	}
}

// handleConnect sends a message back to the conneting client to confirm
// their access and set their initial channel.
func handleConnect(msg *chat.Msg) {
	toId := msg.FromId
	userMap[toId] = msg.From

	// reply
	msg.ChannelId = 0
	msg.From = ""
	msg.ToId = toId

	log.Debug("%+v", msg)
	msgProc.SendMsg(toId, msg)

	// send welcome message
	sendWelcomeMsg(toId, userMap[toId])
}

// handleJoinChan adds a user to a channel, if not already presents and they
// have required rights, and then calls accounceUser().
func handleJoinChan(msg *chat.Msg) {
	ch := createChannel(strings.TrimSpace(msg.Text))
	userId := msg.FromId

	con := ch.GetConnection(userId)
	if con == nil {
		ch.AddConnection(proto.GetConnection(userId))
		announceUserJoin(ch, userId)
	}

	// reply
	msg.ChannelId = ch.Id()
	msg.From = ch.Name()
	msg.ToId = userId

	log.Debug("%+v", msg)
	msgProc.SendMsg(msg.ToId, msg)
}

// handleSetName links a client network ID to a friendly username. If the client
// is registering for the first time, a welcome message is sent back to that
// client.
func handleSetName(msg *chat.Msg) {
	userMap[msg.FromId] = msg.From
	sendSetNameConfirm(msg)
}

// sendJoinChanConfirm sends a message to the client to let it know that it has
// successfully joined a new channel.
func sendJoinChanConfirm(ch *net.BroadcastGroup, toId uint32) {
	msg := new(chat.Msg)
	msg.ChannelId = ch.Id()
	msg.From = ch.Name()
	msg.Subtype = chat.MSG_SUB_JOIN_CHANNEL
	msg.ToId = toId

	log.Debug("%+v", msg)
	msgProc.SendMsg(toId, msg)
}

// sendSetNameConfirm sends a confirmation that the user's name was set
// successfully.
func sendSetNameConfirm(msg *chat.Msg) {
	msg.Text = fmt.Sprintf("Name set to %v", msg.From)
	msg.From = ""
	msg.ToId = msg.FromId

	log.Debug("%+v", msg)
	msgProc.SendMsg(msg.ToId, msg)
}

// sendWelcomeMsg sends a quick hello to newly connected users.
func sendWelcomeMsg(id uint32, name string) {
	conMsg := new(chat.Msg)
	conMsg.ChannelId = 0
	conMsg.Subtype = chat.MSG_SUB_CHAT
	conMsg.ToId = id
	conMsg.Text = fmt.Sprintf("Welcome %v!", name)

	log.Debug("%+v", conMsg)
	msgProc.SendMsg(conMsg.ToId, conMsg)
}
