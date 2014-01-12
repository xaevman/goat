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

// Stdlib imports.
import (
	"sync"
)

// Message header constants.
const (
	HEADER_LEN_B    = 4
	MAX_MSG_TYPE    = 1023
	MAX_NET_MSG_LEN = 32 * 1024
)

// Message flag bitwise offsets.
const (
	msgCompressedOffset = 11
	msgEncryptedOffset  = 12
)

// Msg flag masks.
const (
	msgTypeMask  = 0xFC00
	msgFlagsMask = 0x03FF
)

// Network id pool.
var netId uint32

// Routing map and synchronization.
var (
	routeMutex sync.RWMutex	
	routeMap   map[uint16]*TcpProtocol
)

// CompressionProvider specifies the interface which network protocols will
// use to compress/decompress network messages. All outgoing messages flow
// through Compress(). Only messages received with the compression header bit
// set will flow through Decompress().
type CompressionProvider interface {
	Compress(msg *NetMsg) error
	Decompress(msg *NetMsg) error
}

// CryptoProvider specifies the interface which network protocols will use
// to encrypt and decrypt network messages. All outgoing messages flow
// through Encrypt(). Only messages received with the encrypted header bit
// set will flow through Decrypt().
type CryptoProvider interface {
	Decrypt(msg *NetMsg) error
	Encrypt(msg *NetMsg) error
}

// AccessProvider specifies the interface which network protocols will use
// to authorize messages for sending or processing. All incoming and outgoing
// messages flow through Authorize() and are immediately dropped if it returns
// false.
type AccessProvider interface {
	Authorize(msg *NetMsg) (bool, error)
}

// MsgProcessor specifies the entry and exit points of the network system which
// network protcols use to accept and distribute incoming messages as well as
// accept and disseminate outgoing messages to the correct endpoints.
type MsgProcessor interface {
	ProcessMsg(msg *NetMsg) error
	SendMsg(id uint32, msg *NetMsg)
}

// GetMsgCompressedFlag retrieves bit 11 of the message header, which is used
// to specify whether the message data itself is compressed or not.
func GetMsgCompressedFlag(header uint16) bool {
	return (header & (1 << msgCompressedOffset)) != 0
}

// GetMsgEncryptedFlag retrieves bit 12 of the message header, which is used
// to specify whether the message data itself is encrypted or not.
func GetMsgEncryptedFlag(header uint16) bool {
	return (header & (1 << msgEncryptedOffset)) != 0
}

// GetMsgHeader retrieves the first 2 byte header of raw line data. 
// Packed into the header is the message type signature and flags specifying 
// whether the data payload is compressed and/or encrypted.
func GetMsgHeader(msgData []byte) uint16 {
	if len(msgData) < 2 {
		panic("msgData buffer less than 2 bytes")
	}

	header := uint16(msgData[0]) << 8 | uint16(msgData[1])
	
	return header
}

// GetMsgPayload returns the payload portion of a raw message buffer.
func GetMsgPayload(msgData []byte) []byte {
  	if len(msgData) < 4 {
     	panic("msgData buffer less than 2 bytes")
   	}

   	size := GetMsgSize(msgData)
 
   	return msgData[4:size + 4]
 }

// GetMsgSig retrieves the message type signature out of a raw message header.
func GetMsgSig(header uint16) uint16 {
	sig := header &^ msgTypeMask

	return sig
}

// GetMsgSize retrieves the data size property from raw line data.
func GetMsgSize(msgData []byte) uint16 {
	if len(msgData) < 4 {
		panic("msgData buffer less than 4 bytes")
	}

	size := uint16(msgData[2]) << 8 | uint16(msgData[3])

	return size
}

// RegisterTcpProtocol registers and maps a message type signature to a
// TcpProtocol which will be used to process messages of that type.
func RegisterTcpProtocol(sig uint16, proto *TcpProtocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	routeMap[sig] = proto
}

// SetMsgCompressedFlag sets bit 11 of a raw header object, which is used to
// specify whether the following data block is compressed or not.
func SetMsgCompressedFlag(header *uint16, val bool) {
	if val {
		*header = *header | (1 << msgCompressedOffset)
	} else {
		*header = *header &^ (1 << msgCompressedOffset)
	}
}

// SetMsgEncryptedFlag sets bit 12 of a raw header object, which is used to
// specify whether the following data block is encrypted or not.
func SetMsgEncryptedFlag(header *uint16, val bool) {
	if val {
		*header = *header | (1 << msgEncryptedOffset)
	} else {
		*header = *header &^ (1 << msgEncryptedOffset)
	}
}

// SetMsgHeader sets the first two bytes of a raw data buffer with the supplied
// header.
func SetMsgHeader(header uint16, msgData []byte) {
	if len(msgData) < 2 {
		panic("msgData buffer less than 2 bytes")
	}

	msgData[0] = byte(header >> 8) 
	msgData[1] = byte(header) 
}

// SetMsgPayload takes the supplied message payload, sets the message size
// property and also copies the message payload into the raw buffer.
func SetMsgPayload(data, msgData []byte) {
	SetMsgSize(len(data), msgData)
	copy(msgData[4:], data)
}

// SetMsgSize sets the message size property on a raw data buffer.
func SetMsgSize(size int, msgData []byte) {
	if len(msgData) < 4 {
		panic("msgData buffer less than 4 bytes")
	}

	msgData[2] = byte(size >> 8)
	msgData[3] = byte(size)
}

// UnregisterTcpProtocol removes a mapping between a message type signature
// and the supplied TcpProtocol if such a mapping exists.
func UnregisterTcpProtocol(sig uint16, proto *TcpProtocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	if routeMap[sig] == proto {
		delete(routeMap, sig)
	}
}

// ValidateMsgHeader does some simple validation of the header in a raw
// data buffer.
func ValidateMsgHeader(msgData []byte) bool {
	return len(msgData) > 3
}

// routeMsg takes an incoming NetMsg and routes it to the appropriate protocol
// if one is registered in the route map, otherwise the message is dropped.
func routeMsg(msg *NetMsg) {
	sig := GetMsgSig(msg.header)

	routeMutex.RLock()
	defer routeMutex.RUnlock()

	proto := routeMap[sig]
	if proto == nil {
		return
	}

	proto.rcvMsg(msg)
}

