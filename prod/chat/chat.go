//  ---------------------------------------------------------------------------
//
//  chat.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package chat defines the various message handlers and network objects used
// in the test chat server and client applications.
package chat

// External imports.
import (
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/buffer"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/prod"	
)

// Stdlin imports.
import (
	"errors"
)

// Chat module name.
const CHAT_MOD_NAME = "Chat"

// Perf counters.
const (
	PERF_CHAT_ERR_BAD_OBJ = iota
	PERF_CHAT_ERR_CON_NIL
	PERF_CHAT_ERR_DESERIALIZE
	PERF_CHAT_RCV
	PERF_CHAT_SEND
	PERF_CHAT_COUNT
)

// Perf counter friendly names
var perfChatNames = []string {
	"ErrorBadObject",
	"ErrorConnectionNil",
	"ErrorDeserializeFailed",
	"MessageReceived",
	"MessageSent",
}

// Perf counters.
var chatPerfs = perf.NewCounterSet(
	"Chat",
	PERF_CHAT_COUNT,
	perfChatNames,
)

// Timeout value for sends.
const CHAT_MSG_SEND_TIMEOUT_SEC = 5

// Msg subtypes
const (
	MSG_SUB_CHAT = iota
	MSG_SUB_CMD
	MSG_SUB_CONNECT
	MSG_SUB_JOIN_CHANNEL
	MSG_SUB_LEAVE_CHANNEL
	MSG_SUB_SET_NAME
)

// Channel constants.
const (
	PUB_CHANNEL = "public"
)

// Network protocol object.
var Protocol = net.NewProtocol(CHAT_MOD_NAME)

// Common error messages.
var (
	errConNil = errors.New("net.Msg.Connection() is nil")
)


// Msg represents a message sent from one client to the chat system.
// ChannelId is the channel that the message is being sent to. From is the 
// friendly name of the user sending the message. FromId is the NetID of
// the message sender. Subtype denotes the type of chat message being sent.
// ToID is the NetID of the message recipient. Text is the actual text of 
// the message.
type Msg struct {
	Access    byte
	ChannelId uint32
	From      string
	FromId    uint32
	Subtype   byte
	ToId      uint32
	Text      string
}

// DeserializeMsg takes serialized data and creates a new Msg object
// from the data.
func DeserializeMsg(data []byte) *Msg {
	var cursor int = 0

	var err error
	msg := new(Msg)

	msg.ChannelId, err = buffer.ReadUint32(data, &cursor)
	if err != nil { return nil }

	msg.From, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil }
	
	msg.FromId, err = buffer.ReadUint32(data, &cursor)
	if err != nil { return nil }
	
	msg.Subtype, err = buffer.ReadByte(data, &cursor)
	if err != nil { return nil }
	
	msg.ToId, err = buffer.ReadUint32(data, &cursor)
	if err != nil { return nil }
	
	msg.Text, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil }

	return msg
}

// SerializeMsg converts a Msg object into a serialized stream of bytes.
func SerializeMsg(msg *Msg) []byte {
	var cursor int = 0

	dataLen := 
		buffer.LenUint32()   		+
		buffer.LenString(msg.From) 	+
		buffer.LenUint32() 			+
		buffer.LenByte() 			+
		buffer.LenUint32() 			+
		buffer.LenString(msg.Text)

	data := make([]byte, dataLen)

	buffer.WriteUint32(msg.ChannelId, data, &cursor)
	buffer.WriteString(msg.From, data, &cursor)
	buffer.WriteUint32(msg.FromId, data, &cursor)
	buffer.WriteByte(msg.Subtype, data, &cursor)
	buffer.WriteUint32(msg.ToId, data, &cursor)
	buffer.WriteString(msg.Text, data, &cursor)

	return data
}


// MsgHandler handles the network IO operations surrounding a
// Msg object.
type MsgHandler struct {
	parent  *net.Protocol
	rcvChan chan *Msg
}

// Close is a no-op for MsgHandler
func (this *MsgHandler) Close() {}

// Init saves the reference to the parent protocol and initializes the
// receive channel for use.
func (this *MsgHandler) Init(proto *net.Protocol) {
	this.parent  = proto
	this.rcvChan = make(chan *Msg, 0)
}

// QueryReceiveMsg returns a read-only reference to the receive channel.
// Pointers to new Msg objects that have been received and decoded arrive
// on this channel.
func (this *MsgHandler) QueryReceiveMsg() <-chan *Msg {
	return this.rcvChan
}

// ReceiveMsg is called by the chat protocol when raw data is received
// on the line. New Msg objects, once built, are sent to the receive channel
// which can be queried via QueryReceiveMsg().
func (this *MsgHandler) ReceiveMsg(msg *net.Msg, access byte) error {
	chatMsg := DeserializeMsg(msg.GetPayload())
	if chatMsg == nil {
		chatPerfs.Increment(PERF_CHAT_ERR_DESERIALIZE)
		return net.ErrDeserializeFailed
	}

	con := msg.Connection()
	if con == nil {
		chatPerfs.Increment(PERF_CHAT_ERR_CON_NIL)
		return errConNil
	}

	chatMsg.Access = access
	chatMsg.FromId = con.Id()

	chatPerfs.Increment(PERF_CHAT_RCV)

	this.rcvChan<- chatMsg

	return nil
}

// SendMsg serializes a Msg object and sends it to the requested network
// ID.
func (this *MsgHandler) SendMsg(targetId uint32, data interface{}) error {
	msgObj := data.(*Msg)

	if msgObj == nil {
		chatPerfs.Increment(PERF_CHAT_ERR_BAD_OBJ)
		return net.ErrInvalidType
	}

	dataBuffer := SerializeMsg(msgObj)
	netMsg     := net.NewMsg()
	netMsg.SetMsgType(this.Signature())
	netMsg.SetPayload(dataBuffer)
	netMsg.SetTimeout(CHAT_MSG_SEND_TIMEOUT_SEC)

	chatPerfs.Increment(PERF_CHAT_SEND)

	return this.parent.SendMsg(targetId, netMsg)
}

// Signature returns CHAT_MSG.
func (this *MsgHandler) Signature() uint16 {
	return products.CHAT_MSG
}
