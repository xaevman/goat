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

// Stdlib imports.
import (
	"hash/crc32"
)


// Msg represents the baseline structure of data used for packaging
// network messages to be sent via the net service.
type Msg struct {
	con        Connection
	cursor     int
	data       []byte
	hdrBuffer  []byte
	header     uint64
	timeoutSec int
}

// NewMsg initializes a new Msg object and returns a pointer to it for use.
func NewMsg() *Msg {
	newMsg := Msg {
		con        : nil,
		cursor     : 0,
		data       : nil,
		header     : 0,
		hdrBuffer  : make([]byte, HEADER_LEN_B),
		timeoutSec : DEFAULT_TIMEOUT_SEC,
	}

	return &newMsg
}

// Connection returns the Connection object (if any) associated with this
// Msg object.
func (this *Msg) Connection() Connection {
	return this.con
}

// GetBytes retreives the Msg, fully serialized with header and
// payload, for transmission.
func (this *Msg) GetBytes() []byte {
	buffer := make([]byte, len(this.data) + HEADER_LEN_B)

	SetMsgHeader(this.header, buffer)
	SetMsgPayload(this.data, buffer)

	return buffer
}

// GetHeader retrieves the header portion of the Msg object.
func (this *Msg) GetHeader() uint64 {
	return this.header
}

// GetPayload retrieves the payload portion of the Msg object.
func (this *Msg) GetPayload() []byte {
	return this.data
}

// SetConnection sets the connection associated with this msg.
func (this *Msg) SetConnection(parentCon Connection) {
	this.con = parentCon
}

// SetCompressed sets the compressed flag in this message's header.
func (this *Msg) SetCompressed(value bool) {
	SetMsgCompressedFlag(&this.header, value)
}

// SetEncrypted sets the encryption flag in this message's header.
func (this *Msg) SetEncrypted(value bool) {
	SetMsgEncryptedFlag(&this.header, value)
}

// SetHeader sets the 64bit header for this message.
func (this *Msg) SetHeader(header uint64) {
	this.header = header
}

// SetMsgType sets the message signature in this message's header.
func (this *Msg) SetMsgType(msgType uint16) {
	SetMsgSig(&this.header, msgType)
}

// SetPayload sets the payload buffer for this message and recalculates
// the msg size and checksum.
func (this *Msg) SetPayload(data []byte) {
	this.data = data
}

// SetTimeout sets this message's timeout (in seconds).
func (this *Msg) SetTimeout(timeoutSec int) {
	this.timeoutSec = timeoutSec
}

// TimeoutSec returns the current timeout (in seconds) specified for this
// Msg object.
func (this *Msg) TimeoutSec() int {
	return this.timeoutSec
}

// addData takes bytes off of the line and adds them into the data buffer.
// Once the data buffer is full, any remnants are returned (because they are
// a part of the next message coming in the stream).
func (this *Msg) addData(newData []byte) ([]byte, bool) {
	var dataCount, hdrCount int

	// build header
	if this.cursor < HEADER_LEN_B {
		hdrCount     = copy(this.hdrBuffer[this.cursor:], newData)
		this.cursor += hdrCount

		if this.cursor < HEADER_LEN_B {
			// header not yet complete
			return nil, false
		}

		// header complete, intialize the rest of the object
		this.header    = GetMsgHeader(this.hdrBuffer)
		this.data      = make([]byte, GetMsgSize(this.header))
		this.hdrBuffer = nil
	}

	// fill data buffer
	dataCount = copy(
		this.data[this.cursor - HEADER_LEN_B:],
		newData[hdrCount:],
	)

	this.cursor += dataCount

	// return
	if dataCount + hdrCount < len(newData) {
		// msg complete, data left over
		return newData[dataCount + hdrCount:], true
	}

	if this.cursor > len(this.data) + HEADER_LEN_B {
		panic("net.Msg buffer overflow (cursor > buffer length)")
	}

	// message complete, no data left over
	if this.cursor == len(this.data) + HEADER_LEN_B {
		return nil, true
	}

	// message not yet complete
	return nil, false
}

// isValid computes the checksum on received payload data and compares it
// to the checksum transmitted in the message header. Returns true if the
// checksums match, and false if not.
func (this *Msg) isValid() bool {
	return GetMsgChecksum(this.header) == crc32.ChecksumIEEE(this.data)
}
