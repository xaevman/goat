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

package net

import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/lib/str"
)

import (
	"hash/crc32"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

// Message type IDs.
const (
	PING_MSG_TYPE = 25
	PONG_MSG_TYPE = 26
)

// Header values for HeaderOps test.
var headerTests = []uint16{
	0, 1, 4, 31, 64, 501, 1002, 1023,
}

// Shared data for TCP test.
var (
	msgCount    uint64 = 0
	pingMsgProc        = new(PingMsgProc)
	pingTxt            = "ping"
	pongMsgProc        = new(PongMsgProc)
	pongTxt            = "pong"
	proto              = NewProtocol("TestProto")
	srvAddr            = "127.0.0.1:6600"
)

// Ping message processor
type PingMsgProc struct {
	parent *Protocol
}

func (this *PingMsgProc) Close() {}
func (this *PingMsgProc) Init(proto *Protocol) {
	this.parent = proto
}
func (this *PingMsgProc) ReceiveMsg(msg *Msg, access byte) error {
	log.Debug(
		"[%v->%v]: %v",
		msg.Connection().RemoteAddr(),
		msg.Connection().LocalAddr(),
		string(msg.GetPayload()),
	)

	// send a pong message back to client
	return pongMsgProc.SendMsg(msg.Connection().Id(), pongTxt)

	return nil
}
func (this *PingMsgProc) SendMsg(targetId uint32, data interface{}) error {
	txt := data.(string)
	msg := NewMsg()
	msg.SetMsgType(this.Signature())
	msg.SetPayload([]byte(txt))

	err := this.parent.sendMsg(targetId, msg)
	return err
}
func (this *PingMsgProc) Signature() uint16 {
	return PING_MSG_TYPE
}

// Pong message processor
type PongMsgProc struct {
	parent *Protocol
}

func (this *PongMsgProc) Close() {}
func (this *PongMsgProc) Init(proto *Protocol) {
	this.parent = proto
}
func (this *PongMsgProc) ReceiveMsg(msg *Msg, access byte) error {
	log.Debug(
		"[%v->%v]: %v",
		msg.Connection().RemoteAddr(),
		msg.Connection().LocalAddr(),
		string(msg.GetPayload()),
	)

	return nil
}
func (this *PongMsgProc) SendMsg(targetId uint32, data interface{}) error {
	txt := data.(string)
	msg := NewMsg()
	msg.SetMsgType(this.Signature())
	msg.SetPayload([]byte(txt))

	err := this.parent.sendMsg(targetId, msg)
	return err
}
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
func TestTCPSrv(t *testing.T) {
	log.Info("")

	//log.DebugLogs = true

	// set up the protocol
	proto.AddSignature(pingMsgProc)
	proto.AddSignature(pongMsgProc)
	proto.SetAccessProvider(new(NoSecurity))

	// fire up the tcp server
	srv := NewTCPSrv()
	srv.Start(srvAddr)

	<-time.After(1 * time.Second)

	// run the tests
	runSimpleTcpTest(1, 1, t)
	runSimpleTcpTest(1, 10, t)
	runSimpleTcpTest(1, 100, t)
	runSimpleTcpTest(10, 1, t)
	runSimpleTcpTest(10, 10, t)
	runSimpleTcpTest(10, 100, t)

	// shut everything down gracefully
	proto.Shutdown()
	srv.Stop()
}

// runSimpleTcpTest spawns the given number of clients and asks them to send the
// supplied number of messages, checking to make sure that the perf totals incrememnt
// properly and do not indicate any failures.
func runSimpleTcpTest(cliCount, sendCount int, t *testing.T) {
	log.Info("SimpleTcpTest (%v clients, %v msg/cli)", cliCount, sendCount)

	proto.perfs.Reset()

	cliList := make([]*TCPCli, 0)

	for i := 0; i < cliCount; i++ {
		cli := NewTCPCli()
		err := cli.Dial(srvAddr)
		if err != nil {
			t.Fatal(err)
			return
		}

		cliList = append(cliList, cli)
	}

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

	perfTotal := proto.perfs.Get(PERF_PROTO_SEND_TOTAL).Value()

	log.Info(
		"%v clients, (%v, %v) messages in %v (%.2f msg/sec)",
		len(cliList),
		msgCount,
		perfTotal,
		runTime,
		float64(perfTotal)/runTime.Seconds(),
	)

	log.Info(perf.DumpString())
}

// runClient connects to the test TCPSrv instance and sends lots of
// messages at semi-random intervals.
func runClient(cli *TCPCli, sendCount int, doneChan chan bool, t *testing.T) {
	for i := 0; i < sendCount; i++ {
		pingMsgProc.SendMsg(cli.Socket().Id(), pingTxt)
		atomic.AddUint64(&msgCount, 1)

		// simulate a normal amount of internet latency
		<-time.After(time.Duration(rand.Intn(5)+15) * time.Millisecond)
	}

	cli.Shutdown()
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
