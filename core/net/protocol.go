//  ---------------------------------------------------------------------------
//
//  protocol.go
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
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	"errors"
	"fmt"
	"sync"
)

// Protocol module name.
const MODULE_PROTOCOL_NAME = "Protocol"

// Perf counters.
const (
	PERF_SEND_COUNT = iota
	PERF_SEND_TOTAL
	PERF_RCV_COUNT
	PERF_RCV_TOTAL
	PERF_MSG_TOTAL
	PERF_COUNT
)

// Perf counter friendly names.
var perfNames = []string {
	"SendCount",
	"SendTotal",
	"ReceiveCount",
	"ReceiveTotal",
	"MsgTotal",
}


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
	perfs      *perf.PerfCounters
}

// NewProtocol is a helper constructor function which creates a newly initialized
// Protocol object, registers it with the net service, and returns a pointer to it
// for use.
func NewProtocol(pName string) *Protocol {
	newProto := Protocol {
		cliMap: make(map[uint32]Connection, 0),
		name:   pName,
		sigMap: make(map[uint16]MsgProcessor, 0),
		perfs:  perf.NewPerfCounters(MODULE_PROTOCOL_NAME, PERF_COUNT, perfNames),
	}

	registerProtocol(&newProto)

	return &newProto
}

// AddSig registers a message type signature and its associated message processing
// object with this protocol.
func (this *Protocol) AddSignature(proc MsgProcessor) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.sigMap[proc.Signature()] != nil {
		log.Error(
			"MsgProcessor already registered (sig: %v), aborting registration",
			proc.Signature(),
		)
		return
	}

	proc.Init(this)

	this.sigMap[proc.Signature()] = proc
	registerSignature(proc.Signature(), this)
}

// DeleteSig removes a message type signature and its associated message processing
// object if one exists.
func (this *Protocol) DeleteSignature(proc MsgProcessor) {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.sigMap[proc.Signature()] != proc {
		log.Error(
			"MsgProcessor registered, but doesn't match the call " +
			"to unregister (sig: %v). Aborting...",
			proc.Signature(),
		)
		return
	}

	proc = this.sigMap[proc.Signature()]
	proc.Close()

	delete(this.sigMap, proc.Signature())
	unregisterSignature(proc.Signature(), this)
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
	provider.Init(this)
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
	provider.Init(this)
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
	provider.Init(this)
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
		go con.Close()
		return 0
	}

	access, err := this.security.Authorize(con)
	if err != nil {
		log.Debug("Error authorizing client %v (%v)", con.Id(), err)
		go con.Close()
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

// SendMsg transmits the supplied message to the target connection Id.
func (this *Protocol) SendMsg(id uint32, msg *Msg) error {
	return this.sendMsg(id, msg)
}

// rcvMsg is the message pipeline for incoming messages. First, the protocol
// is checked to see if a message processor is registered. Next, the registered
// AccessProvider is queried to make sure the message is allowed to pass. Then,
// the message is passed through registered Decryption and Decompression 
// processes if registered and necessary. Finally, the pre-processed message is
// passed to the message processor for final processing.
func (this *Protocol) rcvMsg(msg *Msg) {
	defer this.perfs.Increment(PERF_RCV_TOTAL)
	defer this.perfs.Increment(PERF_MSG_TOTAL)

	access := this.getAccess(msg.Con)
	if access < 1 {
		return
	}

	sig := GetMsgSig(msg.Header)

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	proc := this.sigMap[sig]

	if proc == nil {
		log.Debug("No valid message processor (sig %v). Dropping message", sig)
		go msg.Con.Close()
		return
	}

	if GetMsgEncryptedFlag(msg.Header) {
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

	if GetMsgCompressedFlag(msg.Header) {
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

	err := proc.ReceiveMsg(msg, access)
	if err != nil {
		log.Debug(
			"Error processing message (proto: %s, err: %v)", 
			this.name,
			err,
		)
	}

	this.perfs.Increment(PERF_RCV_COUNT)
}

// sendMsg distributes the given msg to a registerd client with that id,
// if one exists.
func (this *Protocol) sendMsg(id uint32, msg *Msg) error {
	defer this.perfs.Increment(PERF_SEND_TOTAL)
	defer this.perfs.Increment(PERF_MSG_TOTAL)

	this.cliMutex.RLock()
	cli := this.cliMap[id]
	this.cliMutex.RUnlock()

	if cli == nil {
		panic("send")
		return errors.New(fmt.Sprintf(
			"sendMsg failed: Client %v doesn't exist.",
			id,
		))
	}

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	sig := GetMsgSig(msg.Header)
	if this.sigMap[sig] == nil {
		panic("send")
		return errors.New(fmt.Sprintf(
			"Can't send a message for an unregistered message type " +
			"signature (%v)",
			sig,
		))
	}

	if GetMsgCompressedFlag(msg.Header) {
		if this.compressor == nil {
			panic("send")
			return errors.New(
				"Compression bit set, but no CompressionProvider " +
				"registered.",
			)
		}

		err := this.compressor.Compress(msg)
		if err != nil {
			panic("send")
			return errors.New(fmt.Sprintf(
				"Error compressing data: %v", err,
			))
		}
	}

	if GetMsgEncryptedFlag(msg.Header) {
		if this.crypto == nil {
			panic("send")
			return errors.New(
				"Encryption bit set, but no CryptoProvider registered.",
			)
		}

		err := this.crypto.Encrypt(msg)
		if err != nil {
			panic("send")
			return errors.New(fmt.Sprintf(
				"Error encrypting data: %v", err,
			))
		}
	}

	cli.Send(msg.GetBytes())
	
	this.perfs.Increment(PERF_SEND_COUNT)

	return nil
}
