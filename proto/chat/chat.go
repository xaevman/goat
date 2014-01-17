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
)

// Stdlin imports.
import (
	"errors"
)

// Chat module name.
const CHAT_MOD_NAME = "Chat"

// Chat service message types.
const (
	CHAT_MSG = iota
)

// Msg subtypes
const (
	MSG_SUB_CHAT = iota
	MSG_SUB_CMD
	MSG_SUB_CONNECT
	MSG_SUB_JOIN_CHANNEL
	MSG_SUB_LEAVE_CHANNEL
	MSG_SUB_SET_NAME
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

// Duplicate creates a copy of the given message and returns a pointer
// to it for use.
func (this *Msg) Duplicate() *Msg {
	newMsg          := new(Msg)
	newMsg.Access    = this.Access
	newMsg.ChannelId = this.ChannelId
	newMsg.From      = this.From
	newMsg.FromId    = this.FromId
	newMsg.Subtype   = this.Subtype
	newMsg.ToId      = this.ToId
	newMsg.Text      = this.Text

	return newMsg
}

// DeserializeMsg takes serialized data and creates a new Msg object
// from the data.
func DeserializeMsg(data []byte) *Msg {
	var cursor int = 0

	msg           := new(Msg)
	msg.ChannelId  = buffer.ReadUint32(data, &cursor)
	msg.From       = buffer.ReadString(data, &cursor)
	msg.FromId     = buffer.ReadUint32(data, &cursor)
	msg.Subtype    = buffer.ReadByte(data, &cursor)
	msg.ToId       = buffer.ReadUint32(data, &cursor)
	msg.Text       = buffer.ReadString(data, &cursor)

	return msg
}

// SerializeMsg converts a Msg object into a serialized stream of bytes.
func SerializeMsg(Msg *Msg) []byte {
	var cursor int = 0

	dataLen := 
		buffer.LenUint32()   		+
		buffer.LenString(Msg.From) 	+
		buffer.LenUint32() 			+
		buffer.LenByte() 			+
		buffer.LenUint32() 			+
		buffer.LenString(Msg.Text)

	data := make([]byte, dataLen)

	buffer.WriteUint32(Msg.ChannelId, data, &cursor)
	buffer.WriteString(Msg.From, data, &cursor)
	buffer.WriteUint32(Msg.FromId, data, &cursor)
	buffer.WriteByte(Msg.Subtype, data, &cursor)
	buffer.WriteUint32(Msg.ToId, data, &cursor)
	buffer.WriteString(Msg.Text, data, &cursor)

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
	chatMsg       := DeserializeMsg(msg.Data)
	chatMsg.Access = access
	chatMsg.FromId = msg.Con.Id()

	this.rcvChan<- chatMsg

	return nil
}

// SendMsg serializes a Msg object and sends it to the requested network
// ID.
func (this *MsgHandler) SendMsg(targetId uint32, data interface{}) error {
	msgObj := data.(*Msg)

	if msgObj == nil {
		return errors.New("Non chat.Msg object received")
	}

	dataBuffer := SerializeMsg(msgObj)
	netMsg     := net.Msg {
		Data:   dataBuffer,
		Header: this.Signature(),
	}

	return this.parent.SendMsg(targetId, &netMsg)
}

// Signature returns CHAT_MSG.
func (this *MsgHandler) Signature() uint16 { 
	return CHAT_MSG 
}
