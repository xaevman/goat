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

// Protocol represents a collection of related clients, message type
// signatures, and the message processing, access, crypto, and compression
// providers which will be used as a part of the messaging pipeline for those
// message types.
type Protocol struct {
	cliMap     map[uint32]Connection
	cliMutex   sync.RWMutex
	compressor CompressionProvider
	crypto     CryptoProvider
	name       string
	objMutex   sync.RWMutex
	security   AccessProvider
	sigMap     map[uint16]MsgProcessor
}

// NewProtocol is a helper constructor function which creates a newly initialized
// Protocol object, registers it with the net service, and returns a pointer to it
// for use.
func NewProtocol(pName string) *Protocol {
	newProto := Protocol {
		cliMap:   make(map[uint32]Connection, 0),
		name:     pName,
		sigMap:   make(map[uint16]MsgProcessor, 0),
	}

	registerProtocol(&newProto)

	return &newProto
}

// AddSig registers a message type signature and its associated message processing
// object with this protocol.
func (this *Protocol) AddSignature(sig uint16, proc MsgProcessor) {
	this.objMutex.Lock()
	this.sigMap[sig] = proc
	this.objMutex.Unlock()

	registerSignature(sig, this)
}

// DeleteSig removes a message type signature and its associated message processing
// object if one exists.
func (this *Protocol) DeleteSignature(sig uint16, proc MsgProcessor) {
	this.objMutex.Lock()
	if this.sigMap[sig] == proc {
		delete(this.sigMap, sig)
	}
	this.objMutex.Unlock()

	unregisterSignature(sig, this)
}

// SetAccessProvider sets the AccessProvider object responsible for authorizing
// messages and clients on this protocol.
func (this *Protocol) SetAccessProvider(provider AccessProvider) {
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
func (this *Protocol) SetCompressionProvider(provider CompressionProvider) {
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
func (this *Protocol) SetCryptoProvider(provider CryptoProvider) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.crypto != nil {
		this.crypto.Close()
	}

	this.crypto = provider
	provider.Init()
}

// Shutdown removes the Protocol from the net service, also unregistering all
// associated message type signatures in the process.
func (this *Protocol) Shutdown() {
	unregisterProtocol(this)
}

// getAccess queries this Protocol's AccessProvider and returns its access level.
// Connections are automatically closed if there is no AccessProvider registered,
// or if an error is returned from the provider during the call to Authorize.
func (this *Protocol) getAccess(con Connection) byte {
	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	if this.security == nil {
		log.Debug(
			"No access provider registered. Dropping client %v", 
			con.Id(),
		)
		con.Close()
		return 0
	}

	access, err := this.security.Authorize(con)
	if err != nil {
		log.Debug("Error authorizing client %v (%v)", con.Id(), err)
		con.Close()
		return 0
	}

	return access
}

// onConnect is notified by the net service of new clients entering the system.
func (this *Protocol) onConnect(con Connection) {
	this.cliMutex.Lock()
	defer this.cliMutex.Unlock()

	access := this.getAccess(con)
	if access < 1 {
		return
	}

	this.cliMap[con.Id()] = con

	log.Debug("Connection %v registered for Proto %v", con.Id(), this.name)
}

// onDisconnect is notified by the net service of clients leaving the system.
func (this *Protocol) onDisconnect(con Connection) {
	this.cliMutex.Lock()
	defer this.cliMutex.Unlock()

	if this.cliMap[con.Id()] == nil {
		return
	}
	
	delete(this.cliMap, con.Id())

	log.Debug("Connection %v unregistered from Proto %v", con.Id(), this.name)
}

// rcvMsg is the message pipeline for incoming messages. First, the protocol
// is checked to see if a message processor is registered. Next, the registered
// AccessProvider is queried to make sure the message is allowed to pass. Then,
// the message is passed through registered Decryption and Decompression 
// processes if registered and necessary. Finally, the pre-processed message is
// passed to the message processor for final processing.
func (this *Protocol) rcvMsg(msg *NetMsg) {
	access := this.getAccess(msg.con)
	if access < 1 {
		return
	}

	sig := GetMsgSig(msg.header)

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	proc := this.sigMap[sig]

	if proc == nil {
		log.Debug("No valid message processor (sig %v). Dropping message", sig)
		msg.con.Close()
		return
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

	err := proc.ProcessMsg(msg, access)
	if err != nil {
		log.Debug(
			"Error processing message. Dropping (proto: %s)", 
			this.name,
		)
	}
}
