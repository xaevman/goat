//  ---------------------------------------------------------------------------
//
//  stdnet.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// External imports.
import (
    "github.com/xaevman/goat/core/log"
)

// Stdlib imports.
import (
    "errors"
    "hash/crc32"
    stdnet "net"
    "net/http"
    "sync"
    "sync/atomic"
)

// Timeouts.
const (
    DEFAULT_EVT_TIMEOUT_SEC = 5
    DEFAULT_MSG_TIMEOUT_SEC = 15
    MAX_TIMEOUT_SEC         = 300
    MIN_TIMEOUT_SEC         = 1
    QUEUE_TIMEOUT_SEC       = 5
)

// Event queue buffer sizes
const QUEUE_BUFFERS = 100

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

// Common error messages.
var (
    ErrDeserializeFailed = errors.New("Deserialization failed")
    ErrInvalidType       = errors.New("Invalid type received")
    ErrBufferTooSmall    = errors.New(
        "msgData buffer not large enough to contain a message",
    )
    ErrInvalidMsgType    = errors.New("msgType > MAX_MSG_TYPE")
    ErrMaxMsgSize        = errors.New("Message size > MAX_NET_MSG_LEN")
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
    LocalAddr() stdnet.Addr
    RemoteAddr() stdnet.Addr
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

// MsgProcessor specifies the interface which user code should implement
// to define the serialization behavior of a given message signature.
type MsgProcessor interface {
    Close()
    DeserializeMsg(msg *Msg, access byte) (interface{}, error)
    Init(proto *Protocol)
    SerializeMsg(data interface{}) (*Msg, error)
    Signature() uint16
}

// NetConnector represents a connector to another network device. Some 
// example implmentations of NetConnector are the built-in TCP and UDP
// client and server objects.
type NetConnector interface {
    Start(addr string) (Connection, error)
    Stop()
}

// EventHandler represents the interface that user code should implement
// to handle events from a given protocol registered in the network layer.
type EventHandler interface {
    Close()
    Init(proto *Protocol)
    OnConnect(con Connection)
    OnDisconnect(con Connection)
    OnError(err error)
    OnReceive(msg interface{}, fromId uint32, access byte)
    OnShutdown()
    OnTimeout(timeout *TimeoutEvent)
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
func GetMsgHeader(msgData []byte) (uint64, error) {
    if len(msgData) < HEADER_LEN_B {
        return 0, ErrBufferTooSmall
    }

    header := uint64(msgData[0]) << 56 | 
        uint64(msgData[1]) << 48 |
        uint64(msgData[2]) << 40 |
        uint64(msgData[3]) << 32 |
        uint64(msgData[4]) << 24 |
        uint64(msgData[5]) << 16 |
        uint64(msgData[6]) << 8  |
        uint64(msgData[7])

    return header, nil
}

// GetMsgPayload returns the payload portion of a raw message buffer.
func GetMsgPayload(msgData []byte) ([]byte, error) {
    if len(msgData) < HEADER_LEN_B {
        return nil, ErrBufferTooSmall
    }

    header, err := GetMsgHeader(msgData)
    if err != nil {
        return nil, err
    }

    size := GetMsgSize(header)

    return msgData[HEADER_LEN_B : HEADER_LEN_B + size], nil
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

// InitHttpSrv initializes the http server. Register handlers as is normal
// for the net/http service.
func InitHttpSrv(addr string) {
    go func() {
        err := http.ListenAndServe(addr, nil)
        if err != nil {
            log.Error(err.Error())
        }
    }()
}

// NextNetID retrieves the next available network ID for use.
func NextNetID() uint32 {
    return atomic.AddUint32(&netId, 1)
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
func SetMsgHeader(header uint64, msgData []byte) error {
    if len(msgData) < HEADER_LEN_B {
        return ErrBufferTooSmall
    }

    msgData[0] = byte(header >> 56)
    msgData[1] = byte(header >> 48)
    msgData[2] = byte(header >> 40)
    msgData[3] = byte(header >> 32)
    msgData[4] = byte(header >> 24)
    msgData[5] = byte(header >> 16)
    msgData[6] = byte(header >>  8)
    msgData[7] = byte(header)

    return nil
}

// SetMsgPayload takes the supplied message payload, sets the message size
// property, computes and sets the checksum property, and also copies 
// the message payload into the raw buffer.
func SetMsgPayload(data, msgData []byte) {
    header, err := GetMsgHeader(msgData)
    if err != nil {
        panic("Setting payload on a message without valid header")
    }
    
    err = SetMsgSize(&header, len(data))
    if err != nil {
        panic(err)
    }

    SetMsgChecksum(&header, crc32.ChecksumIEEE(data))

    err = SetMsgHeader(header, msgData)
    if err != nil {
        panic(err)
    }

    copy(msgData[HEADER_LEN_B:], data)
}

// SetMsgSig sets the first 10 bits of a message header with the supplied
// msgType.
func SetMsgSig(header *uint64, msgType uint16) error {
    if msgType > MAX_MSG_TYPE {
        return ErrInvalidMsgType
    }

    *header = *header | uint64(GetMsgSigPart(msgType))

    return nil
}

// SetMsgSize sets the message size property on a raw data buffer.
func SetMsgSize(header *uint64, size int) error {
    if size > MAX_NET_MSG_LEN {
        return ErrMaxMsgSize
    }

    *header = *header | uint64(size) << 16

    return nil
}

// ValidateMsgHeader does some simple validation of the header in a raw
// data buffer.
func ValidateMsgHeader(msgData []byte) bool {
    return len(msgData) >= HEADER_LEN_B
}
