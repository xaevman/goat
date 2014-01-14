//  ---------------------------------------------------------------------------
//
//  netmsg.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// NetMsg represents the baseline structure of data used for packaging
// network messages to be sent via the net service.
type NetMsg struct {
	con       Connection
	cursor    int
	data      []byte
	header    uint16
	hdrBuffer []byte
}

// NewNetMsg is a constructor helper which returns a pointer to a new
// NetMsg object with the given header and payload.
func NewNetMsg(header uint16, payload []byte) *NetMsg {
	newMsg := NetMsg {
		data:   payload,
		header: header,
	}

	return &newMsg
}

// GetBytes retreives the NetMsg, fully serialized with header and
// payload, for transmission.
func (this *NetMsg) GetBytes() []byte {
	buffer := make([]byte, len(this.data) + 4)

	SetMsgHeader(this.header, buffer)
	SetMsgPayload(this.data,  buffer)

	return buffer
}

// GetHeader retrieves the header portion of the NetMsg object.
func (this *NetMsg) GetHeader() uint16 {
	return this.header
}

// GetPayload retrieves the payload portion of the NetMsg object.
func (this *NetMsg) GetPayload() []byte {
	return this.data
}

// addData takes bytes off of the line and adds them into the data buffer.
// Once the data buffer is full, any remnants are returned (because they are
// a part of the next message coming in the stream).
func (this *NetMsg) addData(msgData []byte) ([]byte, bool) {
	var dCount, hCount int

	if this.cursor < HEADER_LEN_B {
		hCount       = copy(this.hdrBuffer[this.cursor:], msgData)
		this.cursor += hCount

		if this.cursor < HEADER_LEN_B {
			return nil, false
		}

		this.header    = GetMsgHeader(this.hdrBuffer)
		this.data      = make([]byte, GetMsgSize(this.hdrBuffer))
		this.hdrBuffer = nil
	}

	dCount = copy(
		this.data[this.cursor - HEADER_LEN_B:], 
		msgData[hCount:],
	)

	this.cursor += dCount

	if dCount + hCount < len(msgData) {
		return msgData[dCount + hCount:], true
	}

	if this.cursor >= len(this.data) {
		return nil, true
	}

	return nil, false
}
