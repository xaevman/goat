//  ---------------------------------------------------------------------------
//
//  dbg.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package dbg implements a protocol, server and client structure that can
// be included in a goat project to enable realtime performance and diagnostic
// monitoring over a TCP connection.
package dbg

// External imports.
import (
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/buffer"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/prod"
)

// Perf counters.
const (
	PERF_DBG_ERR_BAD_OBJ = iota
	PERF_DBG_ERR_DESERIALIZE
	PERF_DBG_RCV
	PERF_DBG_SEND
	PERF_DBG_COUNT
)

// Perf counter friendly names
var perfDbgNames = []string {
	"ErrorBadObject",
	"ErrorDeserializeFailed",
	"MessageReceived",
	"MessageSent",
}

// Perf counters.
var dbgPerfs = perf.NewCounterSet(
	"Debug",
	PERF_DBG_COUNT,
	perfDbgNames,
)

// Network protocol object.
var Protocol *net.Protocol


// CmdMsg represents a command message being sent into the debugging
// system. Access is the authorized level of access passed up from the
// protocol layer. Cmd is the base command being issued to the server.
// Data contains any arguments passed along with the base command.
// CmdMsg objects are re-used for the replies from client to server as
// well. In a reply, the Access field will contain the access level 
// returned from the server. If Access == 0, then an error occured and
// the relevant error message will be contained in the Data field. If
// there was no error the Data field will contain the JSON encoded
// debugging data from the server.
type CmdMsg struct {
	Access byte
	Cmd    string
	Data   string
}

// Deserialize takes serialized data and creates a new CmdMsg object
// from the data.
func Deserialize(data []byte) *CmdMsg {
	var cursor int = 0

	var err error
	msg := new(CmdMsg)

	msg.Cmd, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil }

	msg.Data, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil }

	return msg
}

// Serialize converts a CmdMsg object into a serialized stream of bytes.
func Serialize(msg *CmdMsg) []byte {
	var cursor int = 0

	dataLen := 
		buffer.LenString(msg.Cmd) +
		buffer.LenString(msg.Data)

	data := make([]byte, dataLen)

	buffer.WriteString(msg.Cmd, data, &cursor)
	buffer.WriteString(msg.Data, data, &cursor)

	return data
}


// CmdMsgHandler is a MsgProcessor implementation which handles network IO
// for CmdMsg objects.
type CmdMsgHandler struct {
	parent  *net.Protocol
	rcvChan chan *CmdMsg
}

// Close is unused in the CmdMsgHandler implementation.
func (this *CmdMsgHandler) Close() {}

// Init saves a reference to the parent protocol and initializes the
// receive channel.
func (this *CmdMsgHandler) Init(proto *net.Protocol) {
	this.parent  = proto
	this.rcvChan = make(chan *CmdMsg, 0)
}

// QueryReceiveMsg returns a read-only reference to the channel on which
// newly received CmdMsg objects have arrived.
func (this *CmdMsgHandler) QueryReceiveMsg() <-chan *CmdMsg {
	return this.rcvChan
}

// ReceiveMsg is called by the parent protocol when raw message data is
// received from the network module. User code should not call ReceiveMsg
// directly. Instead, integrate QueryReceiveMsg() into your IO loop to
// receive encoded CmdMsg objects as they are received.
func (this *CmdMsgHandler) ReceiveMsg(msg *net.Msg, access byte) error {
	cmdMsg := Deserialize(msg.GetPayload())
	if cmdMsg == nil {
		dbgPerfs.Increment(PERF_DBG_ERR_DESERIALIZE)
		return net.ErrDeserializeFailed
	}

	cmdMsg.Access = access

	dbgPerfs.Increment(PERF_DBG_RCV)

	this.rcvChan<- cmdMsg

	return nil
}

// SendMsg takes CmdMsg objects, serializes them and sends them through to
// the network layer for transmission.
func (this *CmdMsgHandler) SendMsg(targetId uint32, data interface{}) error {
	sendMsg := data.(*CmdMsg)

	if sendMsg == nil {
		dbgPerfs.Increment(PERF_DBG_ERR_BAD_OBJ)
		return net.ErrInvalidType
	}

	dataBuffer := Serialize(sendMsg)
	netMsg     := net.NewMsg()
	netMsg.SetMsgType(this.Signature())
	netMsg.SetPayload(dataBuffer)

	dbgPerfs.Increment(PERF_DBG_SEND)

	return this.parent.SendMsg(targetId, netMsg)
}

func (this *CmdMsgHandler) Signature() uint16 {
	return products.DBG_MSG
}

