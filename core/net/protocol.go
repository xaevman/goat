//  ---------------------------------------------------------------------------
//
//  protocol.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// Stdlib imports.
import (
	"sync"
)

// TcpProtocol represents a collection of related Tcp clients, message type
// signatures, and the message processing, access, crypto, and compression
// providers which will be used as a part of the messaging pipeline for those
// message types.
type TcpProtocol struct {
	cliMap     map[uint32]*tcpCli
	compressor CompressionProvider
	crypto     CryptoProvider
	listMutex  sync.RWMutex
	name       string
	security   AccessProvider
	sigMap     map[uint16]MsgProcessor
	sigMutex   sync.RWMutex
}

// NewTcpProtocol is a helper constructor function which returns a pointer to a
// new TcpProtocol object.
func NewTcpProtocol(protoName string) *TcpProtocol {
	newProto := TcpProtocol {
		cliMap: make(map[uint32]*tcpCli, 0),
		name:   protoName,
		sigMap: make(map[uint16]MsgProcessor, 0),
	}

	return &newProto
}

// AddSig registers a message type signature and its associated message processing
// object with this protocol.
func (this *TcpProtocol) AddSig(sig uint16, proc MsgProcessor) {
	this.sigMutex.Lock()
	this.sigMap[sig] = proc
	this.sigMutex.Unlock()

	RegisterTcpProtocol(sig, this)
}

// DeleteSig removes a message type signature and its associated message processing
// object if one exists.
func (this *TcpProtocol) DeleteSig(sig uint16, proc MsgProcessor) {
	this.sigMutex.Lock()
	if this.sigMap[sig] == proc {
		delete(this.sigMap, sig)
	}
	this.sigMutex.Unlock()

	UnregisterTcpProtocol(sig, this)
}

// rcvMsg is the message pipeline for incoming messages. First, the protocol
// is checked to see if a message processor is registered. Next, the registered
// AccessProvider is queried to make sure the message is allowed to pass. Then,
// the message is passed through registered Decryption and Decompression 
// processes if registered and necessary. Finally, the pre-processed message is
// passed the message processor for final processing.
func (this *TcpProtocol) rcvMsg(msg *NetMsg) {
	this.sigMutex.RLock()
	proc := this.sigMap[GetMsgSig(msg.header)]
	this.sigMutex.RUnlock()

	if proc == nil {
		// not a valid sig
		return
	}

	if this.security == nil {
		// no security == fail
		return

		success, err := this.security.Authorize(msg)
		if err != nil {
			// error authorizing
			return
		}

		if !success {
			// client not authorized
			return
		}
	}

	if GetMsgEncryptedFlag(msg.header) {
		if this.crypto == nil {
			// encrypted but no encryption provider
			// to handle msg
			return
		}

		err := this.crypto.Decrypt(msg)
		if err != nil {
			// error decrypting
			return
		}
	}

	if GetMsgCompressedFlag(msg.header) {
		if this.compressor == nil {
			// compressed but no compression provider
			// to handle msg
			return
		}

		err := this.compressor.Decompress(msg)
		if err != nil {
			// error decompressing
			return
		}
	}

	err := proc.ProcessMsg(msg)
	if err != nil {
		// error processing msg
	}
}
