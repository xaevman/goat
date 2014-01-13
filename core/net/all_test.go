//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/strutil"
)

import(
	"math/rand"
	"net"
	"testing"
	"time"
)

// Testing msg type.
const TEST_MSG_TYPE = 25

// Shared data buffer used for TCPSrv test.
var data []byte

// Header values to test.
var headerTests = []uint16 {
	0, 1, 4, 31, 64, 501, 1002, 1023,
}

// Test message processor.
type TextTestMsgProc struct {}

// UNUSED
func (this *TextTestMsgProc) Close() {}

// UNUSED
func (this *TextTestMsgProc) Init() {}

// ProcessMsg takes a NetMsg and logs it to the informational log.
func (this *TextTestMsgProc) ProcessMsg(msg *NetMsg, access byte) error {
	log.Info(
		"From [%v]: %v",
		msg.con.Id(),
		string(msg.GetPayload()),
	)
	return nil
}

// UNUSED
func (this *TextTestMsgProc) SendMsg(id uint32, msg *NetMsg, access byte) error {
	return nil
}


// Test access provider.
type TextTestAccess struct {}

// UNUSED
func (this *TextTestAccess) Close() {}

// UNUSED
func (this *TextTestAccess) Init() {}

// Authorize implements no security. It authorizes all clients and messages.
func (this *TextTestAccess) Authorize(con Connection) (byte, error) {
	return 255, nil
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
	log.DebugLogs = true

	// build msg
	msg    := "test msg"
	header := uint16(TEST_MSG_TYPE)
	data    = make([]byte, len(msg) + 4)
	SetMsgHeader(header, data)
	SetMsgPayload([]byte(msg), data)

	// set up the protocol
	proto := NewProtocol("TestTcpProto")
	proto.AddSignature(TEST_MSG_TYPE, new(TextTestMsgProc))
	proto.SetAccessProvider(new(TextTestAccess))

	// fire up the tcp server
	srv := NewTCPSrv()
	srv.Start("127.0.0.1:6600")

	<-time.After(1 * time.Second)

	for i := 0; i < 100; i++ {
		<-time.After(time.Duration(rand.Intn(100)) * time.Millisecond)
		go runClient(t)
	}

	<-time.After(5 * time.Second)

	proto.Shutdown()
	srv.Stop()
}

// runClient connects to the test TCPSrv instance and sends lots of
// messages at semi-random intervals.
func runClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		conn.Write(data)
		<-time.After(time.Duration(rand.Intn(15)) * time.Millisecond)
	}
}

// testHeaders is a generalized function for testing all of the get and set
// header routines.
func testHeaders(i uint16, t *testing.T) {
	log.Info("Testing headers with type %v", i)

	text    := "This is a test message"
	buffer  := make([]byte, len(text) + HEADER_LEN_B)
	msgType := uint16(i)
	header  := msgType

	// flags should be 0, test
	if GetMsgCompressedFlag(msgType) {
		t.Fatal("Compressed flag is set in new header")
	}
	if GetMsgEncryptedFlag(msgType) {
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
	rtHeader  := GetMsgHeader(buffer)
	rtMsgType := GetMsgSig(rtHeader)

	if !ValidateMsgHeader(buffer) {
		t.Fatalf("Buffer failed header validation", buffer)
	}
	if rtHeader != header {
		t.Fatalf("Roundtrip header: %v != %v\n", rtHeader, header)
	}
	if rtMsgType != msgType {
		t.Fatalf("Roundtrip msgType: %v != %v\n", rtMsgType, msgType)
	}
	if !strutil.StrEq(rtPayload, text) {
		t.Fatalf("Roundtrip payload: %v != %v\n", rtPayload, text)
	}

	// set flags back to 0 and test again
	SetMsgEncryptedFlag(&header, false)
	SetMsgCompressedFlag(&header, false)

	if GetMsgCompressedFlag(msgType) {
		t.Fatal("Compressed flag is set in new header")
	}
	if GetMsgEncryptedFlag(msgType) {
		t.Fatal("Encrypted flag is set in new header")
	}

	log.Info("TestHeaderOps[%v]: passed", i)
}

