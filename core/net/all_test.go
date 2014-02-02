//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package net provides abstractions for TCP servers and clients
// which handle massively parallel IO and present a unified interface
// for implementing security and messaging protocols on top of them.
//
// Network message (max length 32KB)
//		flags
//			11: compressed
//			12: encrypted
//			13: reserved
//			14: reserved
//			15: reserved
//			16: reserved
//
// [0-1]     msgtype (bits 1-10 for 1024 unique msg types), flags (bits 11-16)
// [2-3]     msgsize (uint16)
// [4-7]	 crc32 checksum of payload (uint32)
// [8-32767] payload is msg size
package net

// External imports.
import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/str"
)

// Stdlib imports.
import (
	"errors"
	"hash/crc32"
	"math/rand"
	"testing"
	"time"
)

// Message type IDs.
const (
	PING_MSG_TYPE = 25
	PONG_MSG_TYPE = 26
)

// Header values for HeaderOps test.
var headerTests = []uint16 {
	0, 1, 4, 31, 64, 501, 1002, 1023,
}

// Shared data for TCP test.
var (
	cliproto    *Protocol
	pingMsgProc = new(PingMsgProc)
	pingTxt     = "ping"
	pongMsgProc = new(PongMsgProc)
	pongTxt     = "pong"
	srvproto    *Protocol
	srvAddr     = "127.0.0.1:8900"
)

// PPMsg represents a "ping" or "pong" network message.
type PPMsg struct {
	sender     uint32
	senderAddr string
	msgType    string
}

// PPEventHandler is a net.EventHandler implementation for the
// net protocol tests.
type PPEventHandler struct {
	parent *Protocol
	t      *testing.T
}

// Close performs no action for PPEventHandlers.
func (this *PPEventHandler) Close() {}

// Init saves a reference to the given protocol for future use.
func (this *PPEventHandler) Init(proto *Protocol) {
	this.parent = proto
}

// OnConnect logs information about the new connection.
func (this *PPEventHandler) OnConnect(con Connection) {
	log.Info("Connect (%s): %s", this.parent.name, con.RemoteAddr())
}

// OnDisconnect logs information about the newly disconnected connection
// object.
func (this *PPEventHandler) OnDisconnect(con Connection) {
	log.Info("Disconnect (%s): %s", this.parent.name, con.RemoteAddr())
}

// OnTimeout fails the test.
func (this *PPEventHandler) OnTimeout(timeout *TimeoutEvent) {
	this.t.Fatalf("Timeout: %+v", timeout)
}

// OnReceive performs a type assertion on the incoming message, logs
// the message, and - if the received message was a ping - returns a
// pong message back to the sender.
func (this *PPEventHandler) OnReceive(msg interface{}) {
	pingMsg, ok := msg.(*PPMsg)
	if !ok {
	    this.t.Fatalf("unexpected type %T", msg)
	    return
	}

	log.Info(pingMsg.msgType)

	if pingMsg.msgType == pongTxt {
		return
	}

	// pong!
	this.parent.SendMsg(pingMsg.sender, PONG_MSG_TYPE, pingMsg)
}

// OnError fails the test.
func (this *PPEventHandler) OnError(err error) {
	this.t.Fatal(err)
}

// OnShutdown logs the shutdown event.
func (this *PPEventHandler) OnShutdown() {
	log.Info("Shutting down %s", this.parent.name)
}


// PingMsgProc is the message processor which handles serialization
// for ping messages.
type PingMsgProc struct {
	parent *Protocol
}

// Close performs no action in PingMsgProc.
func (this *PingMsgProc) Close() {}

// Init saves a reference to the parent protocol.
func (this *PingMsgProc) Init(proto *Protocol) {
	this.parent = proto
}

// DeserializeMsg turns a received net.Msg into a new PPMsg object.
func (this *PingMsgProc) DeserializeMsg(
	msg    *Msg, 
	access byte,
) (interface{}, error) {
	pingMsg           := new(PPMsg)
	pingMsg.sender     = msg.Connection().Id()
	pingMsg.senderAddr = msg.Connection().RemoteAddr().String()
	pingMsg.msgType    = string(msg.GetPayload())

	return pingMsg, nil
}

// SerializeMsg takes a given PPMsg ping message and turns it into a
// net.Msg for transmission.
func (this *PingMsgProc) SerializeMsg(data interface{}) (*Msg, error) {
	ppMsg, ok := data.(*PPMsg)
	if !ok {
		return nil, errors.New("Not a *PPMsg type")
	}

	ppMsg.msgType = pingTxt
	msg          := NewMsg()

	msg.SetMsgType(this.Signature())
	msg.SetPayload([]byte(ppMsg.msgType))

	return msg, nil
}

// Signature returns PING_MSG_TYPE.
func (this *PingMsgProc) Signature() uint16 {
	return PING_MSG_TYPE
}


// PongMsgProc is the message processor which handles serialization
// for pong messages.
type PongMsgProc struct {
	parent *Protocol
}

// Close performs no actions in PongMsgProc.
func (this *PongMsgProc) Close() {}

// Init stores a reference to the parent protocol.
func (this *PongMsgProc) Init(proto *Protocol) {
	this.parent = proto
}

// DeserializeMsg turns a received net.Msg into a new PPMsg object.
func (this *PongMsgProc) DeserializeMsg(
	msg    *Msg, 
	access byte,
) (interface{}, error) {
	pingMsg           := new(PPMsg)
	pingMsg.sender     = msg.Connection().Id()
	pingMsg.senderAddr = msg.Connection().RemoteAddr().String()
	pingMsg.msgType    = string(msg.GetPayload())

	return pingMsg, nil
}

// SerializeMsg takes a given PPMsg pong message and turns it into a
// net.Msg for transmission.
func (this *PongMsgProc) SerializeMsg(data interface{}) (*Msg, error) {
	ppMsg, ok := data.(*PPMsg)
	if !ok {
		return nil, errors.New("Not a *PPMsg type")
	}

	ppMsg.msgType = pongTxt
	msg          := NewMsg()

	msg.SetMsgType(this.Signature())
	msg.SetPayload([]byte(ppMsg.msgType))

	return msg, nil
}

// Signature returns PONG_MSG_TYPE.
func (this *PongMsgProc) Signature() uint16 {
	return PONG_MSG_TYPE
}


// TestHeaderOpts runs all of the header set and get options on a variety
// of header message type signatures.
func TestHeaderOps(t *testing.T) {
	for i := range headerTests {
		testHeaders(headerTests[i], t)
	}
}

// TestTCPSrv tests the TCPSrv class with a simple text messaging protocol.
// Tests are performed with a variety of client counts and total message 
// counts.
func TestTCPSrv(t *testing.T) {
	log.Info("")

	// log.DebugLogs = true

	srvHandler  := new(PPEventHandler)
	srvHandler.t = t
	srvproto     = NewProtocol("SrvProto", srvHandler)

	// set up the protocols
	srvproto.AddSignature(pingMsgProc)
	srvproto.AddSignature(pongMsgProc)
	srvproto.SetAccessProvider(new(NoSecurity))

	srvproto.ListenTcp(srvAddr)

	<-time.After(1 * time.Second)

	// run the tests
	runSimpleTcpTest(1, 1, t)
	runSimpleTcpTest(1, 10, t)
	runSimpleTcpTest(1, 100, t)
	runSimpleTcpTest(10, 1, t)
	runSimpleTcpTest(10, 10, t)
	runSimpleTcpTest(10, 100, t)
	runSimpleTcpTest(100, 1, t)
	runSimpleTcpTest(100, 10, t)
	runSimpleTcpTest(100, 100, t)

	// shut everything down gracefully
	srvproto.Shutdown()
}

// runSimpleTcpTest spawns the given number of clients and asks them to send the
// supplied number of messages, checking to make sure that the perf totals incrememnt
// properly and do not indicate any failures.
func runSimpleTcpTest(cliCount, sendCount int, t *testing.T) {
	log.Info("SimpleTcpTest (%v clients, %v msg/cli)", cliCount, sendCount)

	cliHandler  := new(PPEventHandler)
	cliHandler.t = t
	cliproto     = NewProtocol("CliProto", cliHandler)

	cliproto.AddSignature(pingMsgProc)
	cliproto.AddSignature(pongMsgProc)
	cliproto.SetAccessProvider(new(NoSecurity))

	for i := 0; i < cliCount; i++ {
		cliproto.DialTcp(srvAddr)
	}

	cliList := cliproto.GetAllConnections()

	log.Info("%v clients connected", len(cliList))
	doneChan := make(chan bool)
	<-time.After(2 * time.Second)

	start := time.Now()

	for i := range cliList {
		go runClient(cliList[i], sendCount, doneChan, t)
	}

	for i := 0; i < len(cliList); i++ {
		<-doneChan
	}

	runTime := time.Since(start)

	<-time.After(1 * time.Second)

	perfTotal := cliproto.perfs.Get(PERF_PROTO_SEND_TOTAL).Value()

	log.Info(
		"%d clients, %d messages in %v (%.2f msg/sec)",
		len(cliList),
		perfTotal,
		runTime,
		float64(perfTotal) / runTime.Seconds(),
	)
}

// runClient connects to the test TCPSrv instance and sends lots of
// messages at semi-random intervals.
func runClient(cli Connection, sendCount int, doneChan chan bool, t *testing.T) {
	for i := 0; i < sendCount; i++ {
		ppMsg := new(PPMsg)
		cliproto.SendMsg(cli.Id(), PING_MSG_TYPE, ppMsg)

		// simulate a normal amount of internet latency
		<-time.After(time.Duration(rand.Intn(5)+15) * time.Millisecond)
	}

	cli.Close()
	doneChan <- true
}

// testHeaders is a generalized function for testing all of the get and set
// header routines.
func testHeaders(msgSig uint16, t *testing.T) {
	log.Info("Testing headers with type %v", msgSig)

	text := "This is a test message"
	msgSize := uint16(len(text))
	buffer := make([]byte, msgSize+HEADER_LEN_B)
	header := uint64(0)

	SetMsgSig(&header, msgSig)

	// flags should be 0, test
	if GetMsgCompressedFlag(header) {
		t.Fatal("Compressed flag is set in new header")
	}
	if GetMsgEncryptedFlag(header) {
		t.Fatal("Encrypted flag is set in new header")
	}

	// set flags to 1 and test
	SetMsgCompressedFlag(&header, true)
	SetMsgEncryptedFlag(&header, true)

	if !GetMsgCompressedFlag(header) {
		t.Fatal("Compressed flag is 0 after being set")
	}
	if !GetMsgEncryptedFlag(header) {
		t.Fatal("Encrypted flag is 0 after being set")
	}

	// round trip and test header, payload, and flags
	SetMsgHeader(header, buffer)
	SetMsgPayload([]byte(text), buffer)
	rtPayload := string(GetMsgPayload(buffer))
	rtHeader := GetMsgHeader(buffer)
	rtMsgSig := GetMsgSig(rtHeader)
	rtMsgSize := GetMsgSize(rtHeader)
	rtChecksum := GetMsgChecksum(rtHeader)
	rtDataHash := crc32.ChecksumIEEE(GetMsgPayload(buffer))

	if !ValidateMsgHeader(buffer) {
		t.Fatalf("Buffer failed header validation", buffer)
	}
	if rtMsgSig != msgSig {
		t.Fatalf("Roundtrip msgSig: %v != %v\n", rtMsgSig, msgSig)
	}
	if rtMsgSize != msgSize {
		t.Fatalf("Roundtrip msgSize: %v != %v\n", rtMsgSize, msgSize)
	}
	if rtChecksum != rtDataHash {
		t.Fatalf("Checksum mismatch : %v != %v\n", rtChecksum, rtDataHash)
	}
	if !str.StrEq(rtPayload, text) {
		t.Fatalf("Roundtrip payload: %v != %v\n", rtPayload, text)
	}

	// set flags back to 0 and test again
	SetMsgEncryptedFlag(&header, false)
	SetMsgCompressedFlag(&header, false)

	if GetMsgCompressedFlag(header) {
		t.Fatal("Compressed flag is set in new header")
	}
	if GetMsgEncryptedFlag(header) {
		t.Fatal("Encrypted flag is set in new header")
	}

	log.Info("TestHeaderOps[%v]: passed", msgSig)
}
