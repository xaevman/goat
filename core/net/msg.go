//  ---------------------------------------------------------------------------
//
//  msg.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// Msg represents the baseline structure of data used for packaging
// network messages to be sent via the net service.
type Msg struct {
	Con       Connection
	Data      []byte
	Header    uint16

	cursor    int
	hdrBuffer []byte
}

// NewNetMsg is a constructor helper which returns a pointer to a new
// Msg object with the given header and payload.
func NewMsg(header uint16, payload []byte) *Msg {
	newMsg := Msg {
		Data:   payload,
		Header: header,
	}

	return &newMsg
}

// GetBytes retreives the Msg, fully serialized with header and
// payload, for transmission.
func (this *Msg) GetBytes() []byte {
	buffer := make([]byte, len(this.Data) + 4)

	SetMsgHeader(this.Header, buffer)
	SetMsgPayload(this.Data,  buffer)

	return buffer
}

// GetHeader retrieves the header portion of the Msg object.
func (this *Msg) GetHeader() uint16 {
	return this.Header
}

// GetPayload retrieves the payload portion of the Msg object.
func (this *Msg) GetPayload() []byte {
	return this.Data
}

// addData takes bytes off of the line and adds them into the data buffer.
// Once the data buffer is full, any remnants are returned (because they are
// a part of the next message coming in the stream).
func (this *Msg) addData(msgData []byte) ([]byte, bool) {
	var dCount, hCount int

	if this.cursor < HEADER_LEN_B {
		hCount       = copy(this.hdrBuffer[this.cursor:], msgData)
		this.cursor += hCount

		if this.cursor < HEADER_LEN_B {
			return nil, false
		}

		this.Header    = GetMsgHeader(this.hdrBuffer)
		this.Data      = make([]byte, GetMsgSize(this.hdrBuffer))
		this.hdrBuffer = nil
	}

	dCount = copy(
		this.Data[this.cursor - HEADER_LEN_B:], 
		msgData[hCount:],
	)

	this.cursor += dCount

	if dCount + hCount < len(msgData) {
		return msgData[dCount + hCount:], true
	}

	if this.cursor >= len(this.Data) {
		return nil, true
	}

	return nil, false
}
