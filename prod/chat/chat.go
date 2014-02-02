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

package chat

// External imports.
import (
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/buffer"
	"github.com/xaevman/goat/prod"
)

// Stdlin imports.
import (
	"errors"
	"fmt"
)

// Chat module name.
const CHAT_MOD_NAME = "Chat"

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
	Text      string
}


// MsgHandler handles the network IO operations surrounding a
// Msg object.
type MsgHandler struct {}

// Close is a no-op for MsgHandler
func (this *MsgHandler) Close() {}

// Init saves the reference to the parent protocol and initializes the
// receive channel for use.
func (this *MsgHandler) Init(proto *net.Protocol) {}

// DeserializeMsg is called by the chat protocol when raw data is received
// on the line.
func (this *MsgHandler) DeserializeMsg(msg *net.Msg, access byte) (interface{}, error) {
	var err error

	cursor  := 0
	data    := msg.GetPayload()
	chatmsg := new(Msg)

	chatmsg.ChannelId, err = buffer.ReadUint32(data, &cursor)
	if err != nil { return nil, err }

	chatmsg.From, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil, err }
	
	chatmsg.Subtype, err = buffer.ReadByte(data, &cursor)
	if err != nil { return nil, err }
		
	chatmsg.Text, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil, err }

	chatmsg.FromId = msg.From()
	chatmsg.Access = access

	return chatmsg, nil
}

// SerializeMsg serializes a Msg object.
func (this *MsgHandler) SerializeMsg(data interface{}) (*net.Msg, error) {
	cursor      := 0
	chatMsg, ok := data.(*Msg)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot serialize type %T", data))
	}

	dataLen := 
		buffer.LenUint32()   		   +
		buffer.LenString(chatMsg.From) +
		buffer.LenByte() 			   +
		buffer.LenString(chatMsg.Text)

	dataBuffer := make([]byte, dataLen)

	buffer.WriteUint32(chatMsg.ChannelId, dataBuffer, &cursor)
	buffer.WriteString(chatMsg.From, dataBuffer, &cursor)
	buffer.WriteByte(chatMsg.Subtype, dataBuffer, &cursor)
	buffer.WriteString(chatMsg.Text, dataBuffer, &cursor)

	msg := net.NewMsg()
	msg.SetTimeout(CHAT_MSG_SEND_TIMEOUT_SEC)
	msg.SetMsgType(this.Signature())
	msg.SetPayload(dataBuffer)

	return msg, nil
}

// Signature returns CHAT_MSG.
func (this *MsgHandler) Signature() uint16 {
	return prod.CHAT_MSG
}
