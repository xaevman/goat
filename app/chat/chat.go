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

const (
	CHAT_MSG = iota
)

type ChatMsg struct {
	ChannelId uint32
	From      string
	FromId    uint32
	ToId      uint32
	Text      string
}


func DeserializeChatMsg(data []byte) *ChatMsg {
	var cursor int = 0

	chatMsg           := new(ChatMsg)
	chatMsg.ChannelId  = buffer.ReadUint32(data, &cursor)
	chatMsg.From       = buffer.ReadString(data, &cursor)
	chatMsg.FromId     = buffer.ReadUint32(data, &cursor)
	chatMsg.ToId       = buffer.ReadUint32(data, &cursor)
	chatMsg.Text       = buffer.ReadString(data, &cursor)

	return chatMsg
}

func SerializeChatMsg(chatMsg *ChatMsg) []byte {
	var cursor int = 0

	dataLen := 
		4 +						// ChannelId
		8 + len(chatMsg.From) + // From
		4 + 					// FromId
		4 + 					// ToId
		8 + len(chatMsg.Text) 	// Text

	data := make([]byte, dataLen)

	buffer.WriteUint32(chatMsg.ChannelId, data, &cursor)
	buffer.WriteString(chatMsg.From, data, &cursor)
	buffer.WriteUint32(chatMsg.FromId, data, &cursor)
	buffer.WriteUint32(chatMsg.ToId, data, &cursor)
	buffer.WriteString(chatMsg.Text, data, &cursor)

	return data
}


type ChatMsgHandler struct {
	parent  *net.Protocol
	rcvChan chan interface{}
}

func (this *ChatMsgHandler) Close() {}

func (this *ChatMsgHandler) Init(proto *net.Protocol) {
	this.parent  = proto
	this.rcvChan = make(chan interface{}, 0)
}

func (this *ChatMsgHandler) QueryReceiveMsg() <-chan interface{} {
	return this.rcvChan
}

func (this *ChatMsgHandler) ReceiveMsg(msg *net.Msg, access byte) error {
	chatMsg := DeserializeChatMsg(msg.Data)

	this.rcvChan<- &chatMsg

	return nil
}

func (this *ChatMsgHandler) SendMsg(targetId uint32, data interface{}) error {
	dataBuffer := SerializeChatMsg(data.(*ChatMsg))
	netMsg     := net.Msg {
		Data:   dataBuffer,
		Header: this.Signature(),
	}

	return this.parent.SendMsg(targetId, &netMsg)
}

func (this *ChatMsgHandler) Signature() uint16 { 
	return CHAT_MSG 
}
