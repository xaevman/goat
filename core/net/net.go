//  ---------------------------------------------------------------------------
//
//  net.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Network message (max length 32KB)
//		flags 
//			11: compressed
//			12: encrypted
//			13: reserved
//			14: reserved
//			15: reserved
//			16: reserved
//
// [0-1]     msgtype (bits 1-10 for 1024 unique msg types), flags (bits 11-16)
// [2-3]     msgsize (uint16)
// [4-32767] payload is msg size - 2
package net


import (

)


const (
	HEADER_LEN_B    = 4
	MAX_MSG_TYPE    = 1023
	MAX_NET_MSG_LEN = 32 * 1024
)

const (
	msgCompressedOffset = 11
	msgEncryptedOffset  = 12
)

const (
	msgTypeMask  = 0xFC00
	msgFlagsMask = 0x03FF
)


var netId uint32


func GetMsgCompressedFlag(header uint16) bool {
	return (header & (1 << msgCompressedOffset)) != 0
}

func GetMsgEncryptedFlag(header uint16) bool {
	return (header & (1 << msgEncryptedOffset)) != 0
}

func GetMsgHeader(msgData []byte) uint16 {
	if len(msgData) < 2 {
		panic("msgData buffer less than 2 bytes")
	}

	header := uint16(msgData[0]) << 8 | uint16(msgData[1])
	
	return header
}

func GetMsgPayload(msgData []byte) []byte {
	if len(msgData) < 4 {
		panic("msgData buffer less than 2 bytes")
	}

	return msgData[4:]
}

func GetMsgSig(header uint16) uint16 {
	sig := header &^ msgTypeMask

	return sig
}

func GetMsgSize(msgData []byte) uint16 {
	if len(msgData) < 4 {
		panic("msgData buffer less than 4 bytes")
	}

	size := uint16(msgData[2]) << 8 | uint16(msgData[3])

	return size
}

func NewMsg(header uint16, compress, encrypt bool, data []byte) []byte {
	buffer := make([]byte, len(data) + HEADER_LEN_B)
	SetMsgCompressedFlag(&header, compress)
	SetMsgEncryptedFlag(&header, encrypt)
	SetMsgHeader(header, buffer)
	SetMsgPayload(data, buffer)

	return buffer
}

func SetMsgCompressedFlag(header *uint16, val bool) {
	if val {
		*header = *header | (1 << msgCompressedOffset)
	} else {
		*header = *header &^ (1 << msgCompressedOffset)
	}
}

func SetMsgEncryptedFlag(header *uint16, val bool) {
	if val {
		*header = *header | (1 << msgEncryptedOffset)
	} else {
		*header = *header &^ (1 << msgEncryptedOffset)
	}
}

func SetMsgHeader(header uint16, msgData []byte) {
	if len(msgData) < 2 {
		panic("msgData buffer less than 2 bytes")
	}

	msgData[0] = byte(header >> 8) 
	msgData[1] = byte(header) 
}

func SetMsgPayload(data, msgData []byte) {
	SetMsgSize(len(data), msgData)
	copy(msgData[4:], data)
}

func SetMsgSize(size int, msgData []byte) {
	if len(msgData) < 4 {
		panic("msgData buffer less than 4 bytes")
	}

	msgData[2] = byte(size >> 8)
	msgData[3] = byte(size)
}
