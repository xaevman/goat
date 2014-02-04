//  ---------------------------------------------------------------------------
//
//  chattestnet.go
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
	"github.com/xaevman/goat/lib/str"
	"github.com/xaevman/goat/prod"
	"github.com/xaevman/goat/prod/chat"
)

// Stdlib imports.
import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Amount of time to wait before transmitting the next test msg.
const TEST_DELAY_MS = 10

// ChatTest is a net.EventHandler implementation which sends
// randomized messages to the chat server and tests that they
// are echoed back correctly.
type ChatTest struct {
	currentChan uint32
	errors      int32
	exeTime     time.Duration
	proto       *net.Protocol
	startTime   time.Time
	srvId       uint32
	success     int32
	testCount   int
	testMap     map[uint32]bool
	testMutex   sync.Mutex
}

// Close performs no operations in ChatTest.
func (this *ChatTest) Close() {}

// Init seeds rand with this instances instance number and sets up the
// chat message handler. It also disables network level security on the
// chat protocol.
func (this *ChatTest) Init(proto *net.Protocol) {
	this.testMap = make(map[uint32]bool, 0)
	this.proto   = proto
	rand.Seed(int64(myIndex))

	this.proto.AddSignature(new(chat.MsgHandler))
	this.proto.SetAccessProvider(new(net.NoSecurity))
}

// GetResults returns the success and error counts, as well as the total
// execution time of the test.
func (this *ChatTest) GetResults() (int, int, time.Duration) {
	errors  := int(atomic.LoadInt32(&this.errors))
	success := int(atomic.LoadInt32(&this.success))

	return success, errors, this.exeTime
}

// OnConnect saves a reference to the server's netID and sends the connect
// message to the server.
func (this *ChatTest) OnConnect(con net.Connection) {
	this.srvId = con.Id()

	log.Info("OnConnect")

	this.sendConnect()
}

// OnDisconnect begins the protocol shutdown process to end the application
// run. It makes the call in a separate go routine to avoid deadlocking
// the protocol layer.
func (this *ChatTest) OnDisconnect(con net.Connection) {
	go func() {
		this.proto.Shutdown()
		log.Error("Disconnected from server")
	}()
}

// OnError passes error received from the network layer on to the logging
// system.
func (this *ChatTest) OnError(err error) {
	log.Error(err.Error())
}

// OnReceive performs a type assertion on incoming messages, then passes
// valid chat messages to their appropriate handler functions based on
// sub-type.
func (this *ChatTest) OnReceive(msg interface{}) {
	chatMsg, ok := msg.(*chat.Msg)
	if !ok {
		log.Error("Invalid message type %T", msg)
		return
	}

	switch chatMsg.Subtype {
	case chat.MSG_SUB_CHAT:
		this.checkChatMsg(chatMsg)
	case chat.MSG_SUB_CMD:
		return // dont care about commands
	case chat.MSG_SUB_CONNECT:
		this.onConnectMsg(chatMsg)
	case chat.MSG_SUB_JOIN_CHANNEL:
		this.onJoinChannelMsg(chatMsg)
	case chat.MSG_SUB_LEAVE_CHANNEL:
		return // not going to leave the test channel
	case chat.MSG_SUB_SET_NAME:
		this.checkChatName(chatMsg)
	}
}

// OnShutdown performs no operations in ChatTest.
func (this *ChatTest) OnShutdown() {
	goapp.Stop()
}

// OnTimeout passes timeout events on to the logging system and increments
// the error counter appropriately. It makes no attempt to retry messages
// that have timed out.
func (this *ChatTest) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error("%+v", timeout)
	this.addError()
}


// addError increments the error counter.
func (this *ChatTest) addError() {
	atomic.AddInt32(&this.errors, 1)
}

// addSuccess increments the success counter.
func (this *ChatTest) addSuccess() {
	atomic.AddInt32(&this.success, 1)
}

// checkChatMsg reads a message received from the server, determines if
// it is a message that was sent by this client, and then validates that
// it matches the message which is outstanding from this client.
func (this *ChatTest) checkChatMsg(msg *chat.Msg) {
	switch {
	case msg.ChannelId == 0:
		return // dont care about system messages
	case msg.ChannelId != this.currentChan:
		log.Error("Message received on wrong channel")
		this.addError()
	default:
		if !str.StrEq(msg.From, myName) {
			return // someone else's spam
		}

		rcvHash := crc32.ChecksumIEEE([]byte(msg.Text))

		this.testMutex.Lock()
		_, exist := this.testMap[rcvHash]
		this.testMutex.Unlock()

		if !exist {
			// from me, but not represented correctly
			this.addError()
			log.Error(
				"Invalid message received from server "+
					"(from: %s, len: %d, hash: %d, text: %s)",
				msg.From,
				len(msg.Text),
				rcvHash,
				msg.Text,
			)

			return
		}

		// everything looks good
		this.addSuccess()

		this.testMutex.Lock()
		this.testMap[rcvHash] = true
		this.testMutex.Unlock()

		log.Info(
			"Msg validated (len: %d, hash: %d)",
			len(msg.Text),
			rcvHash,
		)
	}
}

// checkChatName checks a setName confirmation message from the server
// to make sure that the name was registered correctly.
func (this *ChatTest) checkChatName(msg *chat.Msg) {
	if msg.Text != fmt.Sprintf("Name set to %v", myName) {
		log.Error("Name incorrect: %s vs %s", msg.Text, myName)
		return
	}

	log.Info("Chat name verified")
}

// endTest is called after the specified number of tests have completed
// and starts initiating the shutdown sequence.
func (this *ChatTest) endTest() {
	log.Info("Test complete")

	this.exeTime = time.Since(this.startTime)

	// give the server some time to respond to the last messages
	<-time.After(10 * time.Second)

	this.OnShutdown()
}

// generateMsg creates a new random string, between 8 and 1024
// characters long and returns it as well as it's computed crc32
// hash.
func (this *ChatTest) generateMsg() (string, uint32) {
	// string length between 8 and 1024 chars
	buffer := make([]byte, rand.Intn(1016)+8)

	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(rand.Intn(94) + 32)
	}

	this.testMutex.Lock()
	defer this.testMutex.Unlock()

	hash := crc32.ChecksumIEEE(buffer)
	this.testMap[hash] = false

	return string(buffer), hash
}

// nextText logs some basic information about the next test that is going
// to be run and sends the selected chat message to the server.
func (this *ChatTest) nextTest() {
	this.testCount++
	msg, hash := this.generateMsg()

	log.Info(
		"Test %d (len: %d, hash: %d)",
		this.testCount,
		len(msg),
		hash,
	)

	this.sendChat(this.currentChan, msg)
}

// onConnectMsg is called when the connect confirmation is received from
// the server. In this case, the client sends a message to join the
// public channel.
func (this *ChatTest) onConnectMsg(msg *chat.Msg) {
	this.sendJoinChannel(chat.PUB_CHANNEL)
}

// onJoinChannelMsg is called when the join channel confirmation is
// received from the server and starts the first test.
func (this *ChatTest) onJoinChannelMsg(msg *chat.Msg) {
	log.Info("Joined channel %d", msg.ChannelId)
	this.currentChan = msg.ChannelId

	this.startTime = time.Now()

	go this.runTests()
}

// runTests loops for the configured number of times and sends randomly
// generated messages to the server.
func (this *ChatTest) runTests() {
	for this.testCount < maxTests {
		<-time.After(TEST_DELAY_MS * time.Millisecond)
		this.nextTest()
	}

	this.endTest()
}

// send transmits a Msg object to the server.
func (this *ChatTest) send(msg *chat.Msg) {
	this.proto.SendMsg(this.srvId, prod.CHAT_MSG, msg)
}

// sendChat sends a chat message to the server.
func (this *ChatTest) sendChat(channel uint32, text string) {
	conMsg          := new(chat.Msg)
	conMsg.ChannelId = channel
	conMsg.From      = myName
	conMsg.Subtype   = chat.MSG_SUB_CHAT
	conMsg.Text      = text

	this.send(conMsg)
}

// sendConnect sends the initial connection message to the server.
func (this *ChatTest) sendConnect() {
	conMsg        := new(chat.Msg)
	conMsg.From    = myName
	conMsg.Subtype = chat.MSG_SUB_CONNECT

	this.send(conMsg)
}

// sendJoinChannel sends a join channel request to the server.
func (this *ChatTest) sendJoinChannel(channelName string) {
	joinMsg        := new(chat.Msg)
	joinMsg.Subtype = chat.MSG_SUB_JOIN_CHANNEL
	joinMsg.Text    = channelName

	this.send(joinMsg)
}
