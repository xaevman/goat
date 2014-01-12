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

// External imports.
import (
	"github.com/xaevman/goat/core/log"
)

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
	cliMutex   sync.RWMutex
	name       string
	objMutex   sync.RWMutex
	security   AccessProvider
	sigMap     map[uint16]MsgProcessor
}

// NewTcpProtocol is a helper constructor function which returns a pointer to a
// new TcpProtocol object.
func NewTcpProtocol(pName string) *TcpProtocol {
	newProto := TcpProtocol {
		cliMap:   make(map[uint32]*tcpCli, 0),
		name:     pName,
		sigMap:   make(map[uint16]MsgProcessor, 0),
	}

	return &newProto
}

// AddSig registers a message type signature and its associated message processing
// object with this protocol.
func (this *TcpProtocol) AddSig(sig uint16, proc MsgProcessor) {
	this.objMutex.Lock()
	this.sigMap[sig] = proc
	this.objMutex.Unlock()

	RegisterTcpProtocol(sig, this)
}

// DeleteSig removes a message type signature and its associated message processing
// object if one exists.
func (this *TcpProtocol) DeleteSig(sig uint16, proc MsgProcessor) {
	this.objMutex.Lock()
	if this.sigMap[sig] == proc {
		delete(this.sigMap, sig)
	}
	this.objMutex.Unlock()

	UnregisterTcpProtocol(sig, this)
}

// SetAccessProvider sets the AccessProvider object responsible for authorizing
// messages and clients on this protocol.
func (this *TcpProtocol) SetAccessProvider(provider AccessProvider) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.security != nil {
		this.security.Close()
	}

	this.security = provider
	provider.Init()
}

// SetCompressionProvider sets the CompressionProvider object responsible for
// handling compression and decompression of messages passing through the
// protocol.
func (this *TcpProtocol) SetCompressionProvider(provider CompressionProvider) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.compressor != nil {
		this.compressor.Close()
	}

	this.compressor = provider
	provider.Init()
}

// SetCryptoProvider sets the CryptoProvider object responsible for
// handling encryption/decryption of messages passing through the protocol.
func (this *TcpProtocol) SetCryptoProvider(provider CryptoProvider) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.crypto != nil {
		this.crypto.Close()
	}

	this.crypto = provider
	provider.Init()
}

// rcvMsg is the message pipeline for incoming messages. First, the protocol
// is checked to see if a message processor is registered. Next, the registered
// AccessProvider is queried to make sure the message is allowed to pass. Then,
// the message is passed through registered Decryption and Decompression 
// processes if registered and necessary. Finally, the pre-processed message is
// passed the message processor for final processing.
func (this *TcpProtocol) rcvMsg(msg *NetMsg) {
	sig := GetMsgSig(msg.header)

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	proc := this.sigMap[sig]

	if proc == nil {
		log.Debug("No valid message processor (sig %v). Dropping message", sig)
		msg.cli.close()
		return
	}

	if this.security == nil {
		log.Debug(
			"No access provider registered. Dropping client %v", 
			msg.cli.id,
		)
		msg.cli.close()
		return

		success, err := this.security.Authorize(msg)
		if err != nil {
			log.Debug("Error authorizing client %v", msg.cli.id)
			msg.cli.close()
			return
		}

		if !success {
			log.Debug("Client not authorized (%v)", msg.cli.id)
			msg.cli.close()
			return
		}
	}

	// register client if not already registered
	this.cliMutex.RLock()
	cli := this.cliMap[msg.cli.id]
	this.cliMutex.RUnlock()

	if cli == nil {
		this.cliMutex.Lock()
		this.cliMap[msg.cli.id] = msg.cli
		this.cliMutex.Unlock()

		log.Debug("Client %v registered for protocol %v", msg.cli.id, this.name)
	}

	if GetMsgEncryptedFlag(msg.header) {
		if this.crypto == nil {
			log.Debug(
				"Encryption flag set, but no encrpytion provider." +
				"Dropping message (proto: %s)",
				this.name,
			)
			return
		}

		err := this.crypto.Decrypt(msg)
		if err != nil {
			log.Debug(
				"Error decrypting message (proto: %s, err: %v)",
				this.name,
				err,
			)
			return
		}
	}

	if GetMsgCompressedFlag(msg.header) {
		if this.compressor == nil {
			log.Debug(
				"Compression flag set, but no compression provider." +
				"Dropping message (proto: %s)",
				this.name,
			)
			return
		}

		err := this.compressor.Decompress(msg)
		if err != nil {
			log.Debug(
				"Error decompressing message (proto: %s, err: %v)",
				this.name,
				err,
			)
			return
		}
	}

	err := proc.ProcessMsg(msg)
	if err != nil {
		log.Debug(
			"Error processing message. Dropping (proto: %s)", 
			this.name,
		)
	}
}
