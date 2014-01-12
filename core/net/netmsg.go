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
	cli    *tcpCli
	cursor int
	data   []byte
	header uint16
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
	count := copy(this.data[this.cursor:], msgData)
	this.cursor += count

	if count < len(msgData) {
		return msgData[count:], true
	}

	if this.cursor == len(this.data) {
		return nil, true
	}

	return nil, false
}
