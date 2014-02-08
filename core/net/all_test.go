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
	"sync"
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
	cliHandler  *PPEventCli
	cliproto    *Protocol
	pingMsgProc = new(PingMsgProc)
	pingTxt     = "ping"
	pongMsgProc = new(PongMsgProc)
	pongTxt     = "pong"
	srvAddr     = "127.0.0.1:8900"
	srvHandler  *PPEventSrv
	srvproto    *Protocol
	waiter      sync.WaitGroup
)


// PPMsg represents a "ping" or "pong" network message.
type PPMsg struct {
	sender     uint32
	senderAddr string
	msgType    string
}


// PPEventSrv is a net.EventHandler implementation for the
// net protocol test server.
type PPEventSrv struct {
	parent     *Protocol
	t          *testing.T
}

// Close performs no action for PPEventSrvs.
func (this *PPEventSrv) Close() {}

// Init saves a reference to the given protocol for future use.
func (this *PPEventSrv) Init(proto *Protocol) {
	this.parent = proto
}

// OnConnect logs information about the new connection.
func (this *PPEventSrv) OnConnect(con Connection) {
	log.Info("Connect (%s): %s", this.parent.name, con.RemoteAddr())
}

// OnDisconnect logs information about the newly disconnected connection
// object.
func (this *PPEventSrv) OnDisconnect(con Connection) {
	log.Info("Disconnect (%s): %s", this.parent.name, con.RemoteAddr())
}

// OnTimeout fails the test.
func (this *PPEventSrv) OnTimeout(timeout *TimeoutEvent) {
	this.t.Fatalf("Timeout: %+v", timeout)
}

// OnReceive performs a type assertion on the incoming message, logs
// the message, and - if the received message was a ping - returns a
// pong message back to the sender.
func (this *PPEventSrv) OnReceive(msg interface{}, fromId uint32, access byte) {
	pingMsg, ok := msg.(*PPMsg)
	if !ok {
	    this.t.Fatalf("unexpected type %T", msg)
	    return
	}

	log.Debug(pingMsg.msgType)

	if pingMsg.msgType == pongTxt {
		return
	}

	// pong!
	this.parent.SendMsg(pingMsg.sender, PONG_MSG_TYPE, pingMsg)
}

// OnError fails the test.
func (this *PPEventSrv) OnError(err error) {
	this.t.Fatalf(err.Error())
}

// OnShutdown logs the shutdown event.
func (this *PPEventSrv) OnShutdown() {
	log.Info("Shutting down %s", this.parent.name)
}


// PPEventCli is a net.EventHandler implementation for the
// net protocol test server.
type PPEventCli struct {
	msgCount int
	parent   *Protocol
	t        *testing.T
}

// Close performs no action for PPEventSrvs.
func (this *PPEventCli) Close() {}

// Init saves a reference to the given protocol for future use.
func (this *PPEventCli) Init(proto *Protocol) {
	this.parent = proto
}

// OnConnect logs information about the new connection.
func (this *PPEventCli) OnConnect(con Connection) {
	log.Info("Connect (%s): %s", this.parent.name, con.RemoteAddr())

	go runTest(con, this.msgCount, this.parent)
}

// OnDisconnect logs information about the newly disconnected connection
// object.
func (this *PPEventCli) OnDisconnect(con Connection) {
	log.Info("Disconnect (%s): %s", this.parent.name, con.RemoteAddr())
}

// OnTimeout fails the test.
func (this *PPEventCli) OnTimeout(timeout *TimeoutEvent) {
	this.t.Fatalf("Timeout: %+v", timeout)
}

// OnReceive performs a type assertion on the incoming message, logs
// the message, and - if the received message was a ping - returns a
// pong message back to the sender.
func (this *PPEventCli) OnReceive(msg interface{}, fromId uint32, access byte) {
	pingMsg, ok := msg.(*PPMsg)
	if !ok {
	    this.t.Fatalf("unexpected type %T", msg)
	    return
	}

	log.Debug(pingMsg.msgType)
}

// OnError fails the test.
func (this *PPEventCli) OnError(err error) {
	this.t.Fatalf(err.Error())
}

// OnShutdown logs the shutdown event.
func (this *PPEventCli) OnShutdown() {
	log.Info("Shutting down %s", this.parent.name)
}

// SetMsgCount sets the target msgCount for this test run.
func (this *PPEventCli) SetMsgCount(count int) {
	this.msgCount = count
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

	// srv
	srvHandler   = new(PPEventSrv)
	srvHandler.t = t
	srvproto     = NewProtocol("SrvProto", srvHandler)
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

// TestUdp performs the same ping/pong network test as runSimpleTcpTest
// but uses UDP as the transport instead.
func TestUdp(t *testing.T) {
	srvAddr := "127.0.0.1:8901"
	cliAddr := "127.0.0.1:8902"

	srvHandler   = new(PPEventSrv)
	srvHandler.t = t
	srvproto     = NewProtocol("UdpSrvTest", srvHandler)
	srvproto.AddSignature(pingMsgProc)
	srvproto.AddSignature(pongMsgProc)
	srvproto.SetAccessProvider(new(NoSecurity))

	_, err := srvproto.ListenUdp(srvAddr)
	if err != nil {
		t.Fatal(err.Error())
	}

	<-time.After(1 * time.Second)

	cliHandler          = new(PPEventCli)
	cliHandler.t        = t
	cliHandler.msgCount = 100
	cliproto            = NewProtocol("UdpCliTest", cliHandler)
	cliproto.AddSignature(pingMsgProc)
	cliproto.AddSignature(pongMsgProc)
	cliproto.SetAccessProvider(new(NoSecurity))

	cliSock, err := cliproto.ListenUdp(cliAddr)
	if err != nil {
		t.Fatal(err.Error())
	}

	<-time.After(1 * time.Second)

	waiter.Add(1)
	err = cliproto.DialUdp(srvAddr, cliSock)
	if err != nil {
		waiter.Done()
		t.Fatal(err.Error())
	}

	log.Info("Transmitting data...")

	waiter.Wait()

	<-time.After(1 * time.Second)

	srvproto.Shutdown()
}

// runSimpleTcpTest spawns the given number of clients and asks them to send the
// supplied number of messages, checking to make sure that the perf totals incrememnt
// properly and do not indicate any failures.
func runSimpleTcpTest(cliCount, sendCount int, t *testing.T) {
	log.Info("SimpleTcpTest (%v clients, %v msg/cli)", cliCount, sendCount)

	cliHandler          = new(PPEventCli)
	cliHandler.t        = t
	cliHandler.msgCount = sendCount
	cliproto            = NewProtocol("CliProto", cliHandler)
	cliproto.AddSignature(pingMsgProc)
	cliproto.AddSignature(pongMsgProc)
	cliproto.SetAccessProvider(new(NoSecurity))

	start := time.Now()

	for i := 0; i < cliCount; i++ {
		waiter.Add(1)
		cliproto.DialTcp(srvAddr)
	}

	log.Info("Transmitting data...")

	waiter.Wait()

	runTime   := time.Since(start)
	perfTotal := cliproto.perfs.Get(PERF_PROTO_SEND_TOTAL).Value()

	log.Info(
		"%d clients, %d messages in %v (%.2f msg/sec)",
		cliCount,
		perfTotal,
		runTime,
		float64(perfTotal) / runTime.Seconds(),
	)

	cliproto.Shutdown()
}

// runTest connects to the test TCPSrv instance and sends lots of
// messages at semi-random intervals.
func runTest(cli Connection, count int, proto *Protocol) {
	<-time.After(1 * time.Second)

	for i := 0; i < count; i++ {
		ppMsg := new(PPMsg)
		proto.SendMsg(cli.Id(), PING_MSG_TYPE, ppMsg)

		// simulate a normal amount of internet latency
		<-time.After(time.Duration(rand.Intn(5)+15) * time.Millisecond)
	}

	<-time.After(1 * time.Second)
	cli.Close()

	waiter.Done()
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
	rtPayloadRaw, err := GetMsgPayload(buffer)
	if err != nil {
		t.Fatal(err)
	}
	rtHeader, err := GetMsgHeader(buffer)
	if err != nil {
		t.Fatal(err)
	}
	rtPayload  := string(rtPayloadRaw)
	rtMsgSig   := GetMsgSig(rtHeader)
	rtMsgSize  := GetMsgSize(rtHeader)
	rtChecksum := GetMsgChecksum(rtHeader)
	rtDataHash := crc32.ChecksumIEEE(rtPayloadRaw)

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
