//  ---------------------------------------------------------------------------
//
//  chattest.go
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
	"github.com/xaevman/goat/proto/chat"
)

// Stdlib imports.
import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)


// ChatTest implements ChatCliNotifier to send a canned set of chat
// messages and verify they are received properly.
type ChatTest struct {
	parent    *chat.ChatCli
	testTimer *time.Timer
}

// OnMsg distributes chat messages to the appropriate handler.
func (this *ChatTest) OnMsg(msg *chat.Msg){
	switch msg.Subtype {
	case chat.MSG_SUB_CHAT:
		this.checkChatMsg(msg)
	case chat.MSG_SUB_CMD:
		return // dont care about commands
	case chat.MSG_SUB_CONNECT:
		this.onConnectMsg(msg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		this.onJoinChannelMsg(msg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return // not going to leave the test channel
	case chat.MSG_SUB_SET_NAME:
		this.checkChatName(msg)
	}
}

// OnConnect is called when a connection to the server is established.
func (this *ChatTest) OnConnect(con net.Connection) {
	log.Info("OnConnect")
	this.parent.SendConnect()
}

// OnDisconnect is called when the connection to the server is dropped.
func (this *ChatTest) OnDisconnect(con net.Connection) {
	go this.parent.Shutdown()

	log.Error("Disconnected from server")

	goapp.Stop()
}

// OnTimeout is called when a TCP timeout is bubbled up from the net
// service. This is counted as an error during the test.
func (this *ChatTest) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error("%+v", timeout)
	errors++
}

// OnShutdown is called when Shutdown() is called on the underlying
// ChatCli object.
func (this *ChatTest) OnShutdown() {
	goapp.Stop()
}

// SetParent is called by a ChatCli object when this object is
// registered with it, in order to give an opportunity to save a reference
// to the parent object.
func (this *ChatTest) SetParent(chatCli *chat.ChatCli) {
	this.parent = chatCli
}

// nextText logs some basic information about the next test that is going
// to be run and sends the selected chat message to the server.
func (this *ChatTest) nextTest() {
	log.Info("Next test (%d)", testCursor)

	if testCount >= maxTests {
		this.endTest()
		return
	}

	this.parent.SendChat(currentChan, myPhrases[testCursor])

	this.testTimer = time.AfterFunc(time.Minute, func() {
		log.Error(
			"Test msg timeout (msg: %d, test: %d)", 
			testCursor,
			testCount,
		)
		errors++
		this.nextCursor()
	})
}

// onConnectMsg is called when the connect confirmation is received from
// the server. In this case, the client sends a message to join the 
// public channel.
func (this *ChatTest) onConnectMsg(msg *chat.Msg) {
	this.parent.SendJoinChannel(chat.PUB_CHANNEL)
}

// onJoinChannelMsg is called when the join channel confirmation is 
// received from the server and starts the first test.
func (this *ChatTest) onJoinChannelMsg(msg *chat.Msg) {
	log.Info("Joined channel %d", msg.ChannelId)
	currentChan = msg.ChannelId
	this.nextTest()
}

// endTest is called after the specified number of tests have completed
// and starts initiating the shutdown sequence.
func (this *ChatTest) endTest() {
	log.Info("Test complete")
	this.OnShutdown()
}

// checkChatMsg reads a message received from the server, determines if
// it is a message that was sent by this client, and then validates that
// it matches the message which is outstanding from this client.
func (this *ChatTest) checkChatMsg(msg *chat.Msg) {
	switch {
	case msg.ChannelId == 0:
		return // dont care about system messages
	case msg.ChannelId != currentChan:
		log.Error("Message received on wrong channel")
		errors++
	default:
		if !strings.Contains(msg.From, myName) {
			return // someone else's spam
		}

		if this.testTimer != nil {
			this.testTimer.Stop()
		}

		if msg.Text != myPhrases[testCursor] {
			// from me, but not represented correctly
			this.nextCursor()
			errors++
			log.Error("%s != %s", msg.Text, myPhrases[testCursor])
			this.nextTest()
			return
		}

		// everything looks good
		this.nextCursor()
		success++
		log.Info("Msg validated (%d)", success)
		this.nextTest()
	}
}

// nextCursor randomly selects the next sentence which should be sent
// to the server.
func (this *ChatTest) nextCursor() {
	testCursor = rand.Intn(len(myPhrases))
	testCount++
}

// checkChatName checks a setName confirmation message from the server
// to make sure that the name was registered correctly.
func (this *ChatTest) checkChatName(msg *chat.Msg) {
	if msg.Text != fmt.Sprintf("Name set to %v", myName) {
		log.Error("Name incorrect: %s vs %s", msg.Text, myName)
		errors++
		return
	}

	log.Info("Chat name verified")
	success++
}

