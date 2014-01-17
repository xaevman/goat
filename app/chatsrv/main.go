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
)

// Network objects.
var (
	discoHandler = net.NewDisconnectChan()
	pubChan 	 = net.NewConnectionGroup("public", net.BROADCAST)
	msgProc 	 = new(chat.MsgHandler)
	proto		 = net.NewProtocol("Chat")
	srv			 = net.NewTCPSrv()
	srvAddr		 = "127.0.0.1:8900"
)

// Synchronization helpers.
var syncObj = lifecycle.New()

// Object maps.
var (
	chanMap = make(map[uint32]*net.ConnectionGroup, 0)
	userMap = make(map[uint32]string, 0)
)

// main is the application entry point
func main() {
	log.DebugLogs = true

	// startup
	proto.AddSignature(msgProc)
	proto.SetAccessProvider(new(net.NoSecurity))
	srv.RegisterDiscoHandler(discoHandler)
	srv.Start(srvAddr)

	proto.RegisterConnection(pubChan)
	chanMap[pubChan.Id()] = pubChan

	// do stuff
	for syncObj.QueryRun() {
		select {
		case msg := <-msgProc.QueryReceiveMsg():
			handleMsg(msg)
		case id := <-discoHandler.QueryDisconnect():
			handleDisco(id)
		case <-syncObj.QueryShutdown():
		}
	}

	// shutdown
	proto.Shutdown()
	srv.Stop()

	syncObj.ShutdownComplete()
}

// handleDisco removes disconnected clients from channel lists.
func handleDisco(id uint32) {
	pubChan.RemoveConnection(id)
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
		sendConnectConfirm(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		return
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return
	case chat.MSG_SUB_SET_NAME:
		registerName(msg)
	}
}


// distChatMsg distributes an incoming chat message to all clients in the given
// channel.
func distChatMsg(msg *chat.Msg) {
	msgProc.SendMsg(msg.ChannelId, msg)
}

// registerName links a client network ID to a friendly username. If the client
// is registering for the first time, a welcome message is sent back to that
// client.
func registerName(msg *chat.Msg) {
	_, exist := userMap[msg.FromId]
	if !exist  {
		defer sendWelcomeMsg(msg.FromId, msg.From)
	}

	userMap[msg.FromId] = msg.From

	sendSetNameConfirm(msg)
}

// sendConnectConfirm sends a message back to the conneting client to confirm
// their access and set their initial channel.
func sendConnectConfirm(msg *chat.Msg) {
	pubChan.AddConnection(proto.GetConnection(msg.FromId))
	sendJoinChanConfirm(pubChan, msg.Duplicate())

	msg.ChannelId = pubChan.Id()
	msg.ToId      = msg.FromId
	msgProc.SendMsg(msg.ToId, msg)
}

// sendJoinChanConfirm sends a message to the client to let it know that it has
// successfully joined a new channel.
func sendJoinChanConfirm(ch *net.ConnectionGroup, msg *chat.Msg) {
	msg.ChannelId = ch.Id()
	msg.From      = ch.Name()
	msg.Subtype   = chat.MSG_SUB_JOIN_CHANNEL
	msg.ToId      = msg.FromId

	msgProc.SendMsg(msg.ToId, msg)
}

// sendSetNameConfirm sends a confirmation that the user's name was set
// successfully.
func sendSetNameConfirm(msg *chat.Msg) {
	msg.Text = fmt.Sprintf("Name set to %v", msg.From)
	msg.From = ""

	msg.ToId = msg.FromId
	msgProc.SendMsg(msg.ToId, msg)
}

// sendWelcomeMsg sends a quick hello to newly connected users.
func sendWelcomeMsg(id uint32, name string) {
	conMsg := chat.Msg {
		From:    "SYSTEM",
		Subtype: chat.MSG_SUB_CHAT,
		ToId:	 id,
		Text:    fmt.Sprintf("Welcome %v!", name),
	}

	msgProc.SendMsg(conMsg.ToId, &conMsg)
}
