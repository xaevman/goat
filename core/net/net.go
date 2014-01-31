//  ---------------------------------------------------------------------------
//
//  net.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package net provides abstractions for TCP servers and clients
// which handle massively parallel IO and present a unified interface
// for implementing security and messaging protocols on top of them.
//
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
// [4-7]	 crc32 checksum of payload (uint32)
// [8-32767] payload is msg size
package net

// External imports.
import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	"errors"
	"hash/crc32"
	"net"
	"sync"
)

// Perf counters.
const (
	PERF_NET_CONNECT = iota
	PERF_NET_DISCONNECT
	PERF_NET_MSG_ROUTE
	PERF_NET_MSG_ROUTE_INVALID
	PERF_NET_PROTO_REGISTER
	PERF_NET_PROTO_UNREGISTER
	PERF_NET_SIG_REGISTER
	PERF_NET_SIG_UNREGISTER
	PERF_NET_TIMEOUT_CONNECT
	PERF_NET_TIMEOUT_DISCONNECT
	PERF_NET_TIMEOUT_GENERAL
	PERF_NET_TIMEOUT_RCV
	PERF_NET_TIMEOUT_SEND
	PERF_NET_COUNT
)

// Perf counters.
var (
	netPerfNames = []string {
		"Connect",
		"Disconnect",
		"MessageRouted",
		"MessageRouteInvalid",
		"ProtocolRegistered",
		"ProtocolUnregistered",
		"SignatureRegistered",
		"SignatureUnregistered",
		"TimeoutConnect",
		"TimeoutDisconnect",
		"TimeoutGeneral",
		"TimeoutReceive",
		"TimeoutSend",
	}
	netPerfs     = perf.NewCounterSet(
		"Module.Net",
		PERF_NET_COUNT,
		netPerfNames,
	)
)

// Timeouts.
const (
	DEFAULT_TIMEOUT_SEC  = 30
	MAX_SEND_TIMEOUT_SEC = 300
	MIN_SEND_TIMEOUT_SEC = 5
)

// Message header constants.
const (
	HEADER_LEN_B    = 8
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

// Network id pool. First 50 Ids are reserved for well-known network
// group objects.
var netId uint32 = 50

// Routing map and synchronization.
var (
	protoMap   = make(map[string]*Protocol)
	routeMutex sync.RWMutex
	sigMap     = make(map[uint16]*Protocol)
)

// Event handler and synchronization.
var (
	eventHandler EventHandler
	eventMutex   sync.RWMutex
)

// Common error messages.
var (
	ErrDeserializeFailed = errors.New("Deserialization failed")
	ErrInvalidType       = errors.New("Invalid type received")
)

// AccessProvider specifies the interface which network protocols will use
// to authorize messages for sending or processing. All incoming messages
// flow through Authorize(). It is left to the individual MsgProcessors to
// decide whether to check authorization for outgoing messages or not.
type AccessProvider interface {
	Authorize(con Connection) (byte, error)
	Close()
	Init(proto *Protocol)
}

// CompressionProvider specifies the interface which network protocols will
// use to compress/decompress network messages. All outgoing messages flow
// through Compress(). Only messages received with the compression header bit
// set will flow through Decompress().
type CompressionProvider interface {
	Close()
	Compress(msg *Msg) error
	Decompress(msg *Msg) error
	Init(proto *Protocol)
}

// Connection specifies the common interface that is used by AccessProvider
// objects to provide authentication for network objects. A given AccessProvider
// may validate based on none, one, or many pieces of the exposed data.
type Connection interface {
	Close()
	Id() uint32
	Key() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Send(data []byte, timeoutSec int)
}

// CryptoProvider specifies the interface which network protocols will use
// to encrypt and decrypt network messages. All outgoing messages flow
// through Encrypt(). Only messages received with the encrypted header bit
// set will flow through Decrypt().
type CryptoProvider interface {
	Close()
	Decrypt(msg *Msg) error
	Encrypt(msg *Msg) error
	Init(proto *Protocol)
}

// EventHandler specifies the interface which TCP clients and servers
// will use to notify subscribers of connect and disconnect events.
type EventHandler interface {
	OnDisconnect(con Connection)
	OnConnect(con Connection)
	OnTimeout(timeout *TimeoutEvent)
}

// MsgProcessor specifies the entry and exit points of the network system which
// network protcols use to accept and distribute incoming messages as well as
// accept and disseminate outgoing messages to the correct endpoints.
type MsgProcessor interface {
	Close()
	Init(proto *Protocol)
	ReceiveMsg(msg *Msg, access byte) error
	SendMsg(targetId uint32, data interface{}) error
	Signature() uint16
}

// GetEventHandler returns the net service's registered EventHandler.
func GetEventHandler() EventHandler {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	return eventHandler
}

// GetMsgCompressedFlag retrieves bit 11 of the message header, which is used
// to specify whether the message data itself is compressed or not.
func GetMsgCompressedFlag(header uint64) bool {
	return (header & (1 << msgCompressedOffset)) != 0
}

// GetMsgEncryptedFlag retrieves bit 12 of the message header, which is used
// to specify whether the message data itself is encrypted or not.
func GetMsgEncryptedFlag(header uint64) bool {
	return (header & (1 << msgEncryptedOffset)) != 0
}

// GetMsgHeader retrieves the 64bit header from a raw message buffer.
func GetMsgHeader(msgData []byte) uint64 {
	if len(msgData) < HEADER_LEN_B {
		panic("msgData buffer less than 8 bytes")
	}

	header := uint64(msgData[0]) << 56 | 
		uint64(msgData[1]) << 48 |
		uint64(msgData[2]) << 40 |
		uint64(msgData[3]) << 32 |
		uint64(msgData[4]) << 24 |
		uint64(msgData[5]) << 16 |
		uint64(msgData[6]) << 8  |
		uint64(msgData[7])

	return header
}

// GetMsgPayload returns the payload portion of a raw message buffer.
func GetMsgPayload(msgData []byte) []byte {
	if len(msgData) < HEADER_LEN_B {
		panic("msgData buffer not large enough to contain a message")
	}

	header := GetMsgHeader(msgData)
	size   := GetMsgSize(header)

	return msgData[HEADER_LEN_B : HEADER_LEN_B + size]
}

// GetMsgSig retrieves the message type signature out of a raw message header.
func GetMsgSig(header uint64) uint16 {
	sig := uint16(header) &^ msgTypeMask

	return sig
}

// GetMsgSigPart returns the message signature portion of a uint16 value.
func GetMsgSigPart(value uint16) uint16 {
	sig := value &^ msgTypeMask
	return sig
}

// GetMsgSize retrieves the data size property from a 64bit header.
func GetMsgSize(header uint64) uint16 {
	size := uint16(header >> 16)

	return size
}

// GetMsgChecksum retrieves the checksum field from raw line data.
func GetMsgChecksum(header uint64) uint32 {
	hash := uint32(header >> 32)

	return hash
}

// SetEventHandler sets the EventHandler object responsible for passing
// connect and disconnect events up from the client and server connection
// layers.
func SetEventHandler(handler EventHandler) {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	eventHandler = handler
}

// SetMsgChecksum sets bytes 4-8 to the computed crc32 hash of the payload
// data.
func SetMsgChecksum(header *uint64, hash uint32) {
	*header = *header | uint64(hash) << 32
}

// SetMsgCompressedFlag sets bit 11 of a raw header object, which is used to
// specify whether the following data block is compressed or not.
func SetMsgCompressedFlag(header *uint64, val bool) {
	if val {
		*header = *header | (1 << msgCompressedOffset)
	} else {
		*header = *header &^ (1 << msgCompressedOffset)
	}
}

// SetMsgEncryptedFlag sets bit 12 of a raw header object, which is used to
// specify whether the following data block is encrypted or not.
func SetMsgEncryptedFlag(header *uint64, val bool) {
	if val {
		*header = *header | (1 << msgEncryptedOffset)
	} else {
		*header = *header &^ (1 << msgEncryptedOffset)
	}
}

// SetMsgHeader sets the first 8 bytes of a raw data buffer with the supplied
// header.
func SetMsgHeader(header uint64, msgData []byte) {
	if len(msgData) < HEADER_LEN_B {
		panic("msgData buffer less than 2 bytes")
	}

	msgData[0] = byte(header >> 56)
	msgData[1] = byte(header >> 48)
	msgData[2] = byte(header >> 40)
	msgData[3] = byte(header >> 32)
	msgData[4] = byte(header >> 24)
	msgData[5] = byte(header >> 16)
	msgData[6] = byte(header >>  8)
	msgData[7] = byte(header)
}

// SetMsgPayload takes the supplied message payload, sets the message size
// property, computes and sets the checksum property, and also copies 
// the message payload into the raw buffer.
func SetMsgPayload(data, msgData []byte) {
	header := GetMsgHeader(msgData)
	SetMsgSize(&header, len(data))
	SetMsgChecksum(&header, crc32.ChecksumIEEE(data))
	SetMsgHeader(header, msgData)

	copy(msgData[HEADER_LEN_B:], data)
}

// SetMsgSig sets the first 10 bits of a message header with the supplied
// msgType.
func SetMsgSig(header *uint64, msgType uint16) {
	if msgType > MAX_MSG_TYPE {
		panic("msgType > MAX_MSG_TYPE")
	}

	*header = *header | uint64(GetMsgSigPart(msgType))
}

// SetMsgSize sets the message size property on a raw data buffer.
func SetMsgSize(header *uint64, size int) {
	if size > MAX_NET_MSG_LEN {
		panic("Message size > MAX_NET_MSG_LEN")
	}

	*header = *header | uint64(size) << 16
}

// ValidateMsgHeader does some simple validation of the header in a raw
// data buffer.
func ValidateMsgHeader(msgData []byte) bool {
	return len(msgData) >= HEADER_LEN_B
}

// onConnect is called by appropriate, connection-based client and server
// objects to notify Protocols of clients coming into, and exiting, the system.
func onConnect(con Connection) {
	netPerfs.Increment(PERF_NET_CONNECT)

	addr, _, _ := net.SplitHostPort(con.RemoteAddr().String())

	log.Debug(
		"%v[%v]->%v connected",
		addr,
		con.Id(),
		con.LocalAddr(),
	)

	routeMutex.RLock()

	for _, proto := range protoMap {
		proto.onConnect(con)
	}

	routeMutex.RUnlock()

	eventMutex.RLock()
	defer eventMutex.RUnlock()

	if eventHandler != nil {
		eventHandler.OnConnect(con)
	}
}

// onDisconnect is called by appropriate, connection-based client and server
// objects to notify Protocols of clients coming into, and exiting, the system.
func onDisconnect(con Connection) {
	netPerfs.Increment(PERF_NET_DISCONNECT)

	addr, _, _ := net.SplitHostPort(con.RemoteAddr().String())

	log.Debug(
		"%v[%v]->%v disconnected",
		addr,
		con.Id(),
		con.LocalAddr(),
	)

	routeMutex.RLock()

	for _, proto := range protoMap {
		proto.onDisconnect(con)
	}

	routeMutex.RUnlock()

	eventMutex.RLock()
	defer eventMutex.RUnlock()

	if eventHandler != nil {
		eventHandler.OnDisconnect(con)
	}
}

// onTimeout calls the existing EventHandler implementation's OnTimeout function
// to notify higher level clients about network timeout events.
func onTimeout(
	timeoutType int,
	messageType uint16,
	parentId    uint32,
	data        interface{},
) {
	switch timeoutType {
	case TIMEOUT_CONNECT:
		netPerfs.Increment(PERF_NET_TIMEOUT_CONNECT)
	case TIMEOUT_DISCONNECT:
		netPerfs.Increment(PERF_NET_TIMEOUT_DISCONNECT)
	case TIMEOUT_GENERAL:
		netPerfs.Increment(PERF_NET_TIMEOUT_GENERAL)
	case TIMEOUT_RCV:
		netPerfs.Increment(PERF_NET_TIMEOUT_RCV)
	case TIMEOUT_SEND:
		netPerfs.Increment(PERF_NET_TIMEOUT_SEND)
	}

	eventMutex.RLock()
	defer eventMutex.RUnlock()

	if eventHandler == nil {
		return
	}

	timeout := TimeoutEvent{
		Data:        data,
		MessageType: messageType,
		ParentId:    parentId,
		TimeoutType: timeoutType,
	}

	eventHandler.OnTimeout(&timeout)
}

// routeMsg takes an incoming Msg and routes it to the appropriate protocol
// if one is registered in the route map, otherwise the message is dropped.
func routeMsg(msg *Msg) {
	sig := GetMsgSig(msg.GetHeader())

	routeMutex.RLock()
	proto := sigMap[sig]
	if proto == nil {
		routeMutex.RUnlock()
		netPerfs.Increment(PERF_NET_MSG_ROUTE_INVALID)
		return
	}
	routeMutex.RUnlock()

	proto.rcvMsg(msg)

	netPerfs.Increment(PERF_NET_MSG_ROUTE)
}

// registerProtocol registers a new Protocol object with the net service.
// This is done automatically when a new Protocol object is created.
func registerProtocol(proto *Protocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	if protoMap[proto.name] != nil {
		log.Error(
			"Protocol %v already registered. Aborting registration",
			proto.name,
		)
		return
	}

	protoMap[proto.name] = proto

	netPerfs.Increment(PERF_NET_PROTO_REGISTER)

	log.Info("Protocol %v registered", proto.name)
}

// registerSignature registers and maps a message type signature to a
// Protocol which will be used to process messages of that type.
func registerSignature(sig uint16, proto *Protocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	if protoMap[proto.name] == nil {
		log.Error(
			"Cannot register a Signature for an "+
				"unregistered Protocol (sig: %v, proto: %v)",
			sig,
			proto.name,
		)

		return
	}

	if protoMap[proto.name] != proto {
		log.Error(
			"Error registering Signature: Specified protocol name "+
				"is already registered, but with a different object "+
				"(sig: %v, proto: %v)",
			sig,
			proto.name,
		)

		return
	}

	if sigMap[sig] != nil {
		log.Error(
			"Signature %v already registered. Aborting registration",
			sig,
		)

		return
	}

	sigMap[sig] = proto

	netPerfs.Increment(PERF_NET_SIG_REGISTER)

	log.Info("Proto %v, Sig %v registered", proto.name, sig)
}

// unregisterProtocol removes a registered protocol from the net service and
// also unregisters any related message type signatures.
func unregisterProtocol(proto *Protocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	if protoMap[proto.name] == proto {
		delete(protoMap, proto.name)
		netPerfs.Increment(PERF_NET_PROTO_UNREGISTER)
		log.Info("Protocol %v unregistered", proto.name)
	}

	for k, v := range sigMap {
		if v == proto {
			delete(sigMap, k)
			netPerfs.Increment(PERF_NET_SIG_UNREGISTER)
			log.Info("Proto %v, Sig %v unregistered", proto.name, k)
		}
	}
}

// unregisterSignature removes a mapping between a message type signature
// and the supplied Protocol if such a mapping exists.
func unregisterSignature(sig uint16, proto *Protocol) {
	routeMutex.Lock()
	defer routeMutex.Unlock()

	if sigMap[sig] == proto {
		delete(sigMap, sig)
		netPerfs.Increment(PERF_NET_SIG_UNREGISTER)
		log.Info("Proto %v, Sig %v unregistered", proto.name, sig)
	}
}
