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
	"github.com/xaevman/goat/lib/lifecycle"
	"github.com/xaevman/goat/lib/math"
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	"errors"
	"fmt"
	"sync"
)

// Perf counters.
const (
	PERF_PROTO_CONNECT = iota
	PERF_PROTO_DISCONNECT
	PERF_PROTO_ERR_AUTH_CLIENT
	PERF_PROTO_ERR_DESERIALIZE
	PERF_PROTO_ERR_NO_ACCESS
	PERF_PROTO_ERR_NO_PROVIDER
	PERF_PROTO_ERR_RCV_CHECKSUM
	PERF_PROTO_ERR_RCV_CON_NIL
	PERF_PROTO_ERR_RCV_DECRYPT
	PERF_PROTO_ERR_RCV_DECOMPRESS
	PERF_PROTO_ERR_SEND_COMPRESS
	PERF_PROTO_ERR_SEND_ENCRYPT
	PERF_PROTO_ERR_SEND_INVALID_CLI
	PERF_PROTO_ERR_SEND_INVALID_MSG_TYPE
	PERF_PROTO_ERR_SERIALIZE
	PERF_PROTO_RCV_BYTES
	PERF_PROTO_RCV_OK
	PERF_PROTO_RCV_TOTAL
	PERF_PROTO_SEND_BYTES
	PERF_PROTO_SEND_OK
	PERF_PROTO_SEND_TOTAL
	PERF_PROTO_TIMEOUT_CONNECT
	PERF_PROTO_TIMEOUT_DISCONNECT
	PERF_PROTO_TIMEOUT_GENERAL
	PERF_PROTO_TIMEOUT_RCV
	PERF_PROTO_TIMEOUT_SEND
	PERF_PROTO_COUNT
)

// Perf counter friendly names.
var protoPerfNames = []string {
	"Connect",
	"Disconnect",
	"ErrorAuthClient",
	"ErrorDeserialize",
	"ErrorNoAccess",
	"ErrorNoProvider",
	"ErrorReceiveChecksum",
	"ErrorReceiveConNil",
	"ErrorReceiveDecrypt",
	"ErrorReceiveDecompress",
	"ErrorSendCompress",
	"ErrorSendEncrypt",
	"ErrorSendInvalidCli",
	"ErrorSendInvalidMsgType",
	"ErrorSerialize",
	"ReceiveBytes",
	"ReceiveSuccess",
	"ReceiveTotal",
	"SendBytes",
	"SendSuccess",
	"SendTotal",
	"ConnectTimeout",
	"DisconnectTimeout",
	"GeneralTimeout",
	"ReceiveTimeout",
	"SendTimeout",
}

// Common error messages.
var (
	errBadChecksum      = errors.New("Malformed message received " + 
		"(checksum mismatch)")
	errConnectionNil    = errors.New("Malformed message received " + 
		"(Connection nil)")
	errNoAccess         = errors.New("Access denied")
	errNoCompProvider   = errors.New("Compression bit set, " +
		"but no CompressionProvider registered.")
	errNoCryptoProvider = errors.New("Encryption bit set, " +
		"but no CryptoProvider registered.")
)

// perfName returns the name to be used for registering with the perf provider,
// given the supplied base name.
func perfName(baseName string) string {
	return fmt.Sprintf("Module.Net.Proto.%s", baseName)
}


// NewProtocol is a helper constructor function which creates a newly initialized
// Protocol object, registers it with the net service, and returns a pointer to it
// for use.
func NewProtocol(pName string, evtHandler EventHandler) *Protocol {
	newProto := Protocol{
		cliMap      : make(map[uint32]Connection, 0),
		connectChan : make(chan Connection, 0),
		discoChan   : make(chan Connection, 0),
		errChan     : make(chan error, 0),
		evtHandler  : evtHandler,
		name        : pName,
		netObjects  : make([]NetConnector, 0),
		perfs       : perf.NewCounterSet(
			perfName(pName), 
			PERF_PROTO_COUNT, 
			protoPerfNames,
		),
		rcvChan     : make(chan *Msg, 0),
		sendChan    : make(chan uint32, 0),
		sigMap      : make(map[uint16]MsgProcessor, 0),
		syncObj     : lifecycle.New(),
		timeoutChan : make(chan *TimeoutEvent, 0),
	}

	newProto.evtHandler.Init(&newProto)
	go newProto.handleEvents()

	return &newProto
}

// Protocol represents a collection of related clients, message type
// signatures, and the message processing, access, crypto, and compression
// providers which will be used as a part of the messaging pipeline for those
// message types.
type Protocol struct {
	cliMap      map[uint32]Connection
	cliMutex    sync.RWMutex
	compressor  CompressionProvider
	connectChan chan Connection
	crypto      CryptoProvider
	discoChan   chan Connection
	errChan     chan error
	evtHandler  EventHandler
	name        string
	netObjects  []NetConnector
	objMutex    sync.RWMutex
	perfs       *perf.CounterSet
	rcvChan     chan *Msg
	security    AccessProvider
	sendChan    chan uint32
	sigMap      map[uint16]MsgProcessor
	syncObj     *lifecycle.Lifecycle
	timeoutChan chan *TimeoutEvent
}

// AddSig registers a message type signature and its associated message processing
// object with this protocol.
func (this *Protocol) AddSignature(proc MsgProcessor) {
	if proc == nil {
		return
	}

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

	log.Info(
		"Signature %d registered in protocol %s", 
		proc.Signature(), 
		this.name,
	)
}

// DeleteSig removes a message type signature and its associated message processing
// object if one exists.
func (this *Protocol) DeleteSignature(proc MsgProcessor) {
	if proc == nil {
		return
	}

	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	if this.sigMap[proc.Signature()] != proc {
		log.Error(
			"MsgProcessor registered, but doesn't match the call "+
				"to unregister (sig: %v). Aborting...",
			proc.Signature(),
		)
		return
	}

	proc = this.sigMap[proc.Signature()]
	proc.Close()

	delete(this.sigMap, proc.Signature())

	log.Info(
		"Signature %d unregistered from protocol %s", 
		proc.Signature(), 
		this.name,
	)
}

// DialTcp attempts to create a tcpCli and connect it to the given
// network address.
func (this *Protocol) DialTcp(addr string) error {
	tcpCli := newtcpCli(this)

	this.objMutex.Lock()
	this.netObjects = append(this.netObjects, tcpCli)
	this.objMutex.Unlock()

	con, err := tcpCli.Start(addr)
	if err != nil {
		return err
	}

	this.onConnect(con)

	return err
}

// GetAllConnections returns a slice containing ll connections associated
// with the protocol object.
func (this *Protocol) GetAllConnections() []Connection {
	this.cliMutex.RLock()
	defer this.cliMutex.RUnlock()

	cursor  := 0
	results := make([]Connection, len(this.cliMap))
	for k, _ := range this.cliMap {
		results[cursor] = this.cliMap[k]
		cursor++
	}

	return results
}

// GetConnection queries the protocol's list of registered connections and
// returns the one matching the supplied NetId, otherwise it returns nil.
func (this *Protocol) GetConnection(id uint32) Connection {
	this.cliMutex.RLock()
	defer this.cliMutex.RUnlock()

	return this.cliMap[id]
}

// ListenTcp attempts to set up a tcpSrv instance listening on the given
// address.
func (this *Protocol) ListenTcp(addr string) error {
	this.objMutex.Lock()
	defer this.objMutex.Unlock()

	tcpSrv         := newtcpSrv(this)
	this.netObjects = append(this.netObjects, tcpSrv)

	_, err := tcpSrv.Start(addr)
	return err
}

// RegisterConnection registers a new connection object with this protocol.
// Connections which attempt to send messages that are a part of this
// protocol will auto-register, but RegisterConnection provides a manual way
// of adding Connection objects or ConnectionGroups as clients of the protocol.
func (this *Protocol) RegisterConnection(con Connection) {
	if con == nil {
		return
	}

	this.cliMutex.Lock()
	defer this.cliMutex.Unlock()

	if this.cliMap[con.Id()] != nil {
		log.Error(
			"Connection already registered (%v), aborting registration",
			con.Id(),
		)
		return
	}

	this.cliMap[con.Id()] = con
}

// SendMsg transmits the supplied message to the target connection Id.
func (this *Protocol) SendMsg(id uint32, sig uint16, msg interface{}) error {
	return this.sendMsg(id, sig, msg)
}

// SetAccessProvider sets the AccessProvider object responsible for authorizing
// messages and clients on this protocol.
func (this *Protocol) SetAccessProvider(provider AccessProvider) {
	if provider == nil {
		return
	}

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
	if provider == nil {
		return
	}

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
	if provider == nil {
		return
	}

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
	this.objMutex.Lock()
	for _, obj := range this.netObjects {
		obj.Stop()
	}
	this.objMutex.Unlock()

	this.syncObj.Shutdown()
	this.evtHandler.Close()
}

// getAccess queries this Protocol's AccessProvider and returns its access level.
// Connections are automatically closed if there is no AccessProvider registered,
// or if an error is returned from the provider during the call to Authorize.
func (this *Protocol) getAccess(con Connection) byte {
	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	if this.security == nil {
		log.Error(
			"No access provider registered. Dropping client %v",
			con.Id(),
		)
		con.Close()
		return 0
	}

	access, err := this.security.Authorize(con)
	if err != nil {
		this.perfs.Increment(PERF_PROTO_ERR_AUTH_CLIENT)
		log.Error("Error authorizing client %v (%v)", con.Id(), err)
		con.Close()
		return 0
	}

	return access
}

// handleEvents is launched within a new go routine when a new protocol is
// instantiated. handleEvents runs continuously and feeds events from the
// protocol layer to the user's registered EventHandler.
func (this *Protocol) handleEvents() {
	for this.syncObj.QueryRun() {
		select {
		case con := <-this.connectChan:
			this.onConnect(con)
		case con := <-this.discoChan:
			this.onDisconnect(con)
		case err := <-this.errChan:
			this.onError(err)
		case msg := <-this.rcvChan:
			this.rcvMsg(msg)
		case timeout := <-this.timeoutChan:
			this.onTimeout(timeout)
		case <-this.syncObj.QueryShutdown():
			this.evtHandler.OnShutdown()
		}
	}

	// unblock this.Shutdown()
	this.syncObj.ShutdownComplete()
}

// onConnect is notified by the net service of new clients entering the system.
func (this *Protocol) onConnect(con Connection) {
	this.getAccess(con)

	this.cliMutex.Lock()
	this.cliMap[con.Id()] = con
	this.cliMutex.Unlock()
	
	this.evtHandler.OnConnect(con)

	this.perfs.Increment(PERF_PROTO_CONNECT)

	log.Debug("Connection %v registered for Proto %v", con.Id(), this.name)
}

// onDisconnect is notified by the net service of clients leaving the system.
func (this *Protocol) onDisconnect(con Connection) {
	this.cliMutex.Lock()
	if this.cliMap[con.Id()] == nil {
		this.cliMutex.Unlock()
		return
	}

	delete(this.cliMap, con.Id())
	this.cliMutex.Unlock()

	this.evtHandler.OnDisconnect(con)

	this.perfs.Increment(PERF_PROTO_DISCONNECT)

	log.Debug("Connection %v unregistered from Proto %v", con.Id(), this.name)
}

// onError is called when error events occur within the protocol.
// onError attempts to feed the registered EventHandler with the
// given data, and itself times out and logs an error after
// DEFAULT_EVT_TIMEOUT_SEC.
func (this *Protocol) onError(err error) {
	this.evtHandler.OnError(err)
}

// onRcv is called when a data receive is completed within the protocol.
// onRcv attempts to feed the registered EventHandler with the given data,
// and itself times out and logs an error after DEFAULT_EVT_TIMEOUT_SEC.
func (this *Protocol) onRcv(msg *Msg) {
	this.rcvMsg(msg)
}

// onTimeout is called when a timeout event bubbles up from the network layer.
// onTimeout attempts to feed the registered EventHandler with the given timeout
// event, and itself times out and logs an error after DEFUALT_EVT_TIMEOUT_SEC.
func (this *Protocol) onTimeout(timeout *TimeoutEvent) {
	switch timeout.TimeoutType {
	case TIMEOUT_CONNECT:
		this.perfs.Increment(PERF_PROTO_TIMEOUT_CONNECT)
	case TIMEOUT_DISCONNECT:
		this.perfs.Increment(PERF_PROTO_TIMEOUT_DISCONNECT)
	case TIMEOUT_GENERAL:
		this.perfs.Increment(PERF_PROTO_TIMEOUT_GENERAL)
	case TIMEOUT_RCV:
		this.perfs.Increment(PERF_PROTO_TIMEOUT_RCV)
	case TIMEOUT_SEND:
		this.perfs.Increment(PERF_PROTO_TIMEOUT_SEND)
	}

	this.evtHandler.OnTimeout(timeout)
}

// rcvMsg is the message pipeline for incoming messages. First, the protocol
// is checked to see if a message processor is registered. Next, the registered
// AccessProvider is queried to make sure the message is allowed to pass. Then,
// the message is passed through registered Decryption and Decompression
// processes if registered and necessary. Finally, the pre-processed message is
// passed to the message processor for final processing.
func (this *Protocol) rcvMsg(msg *Msg) {
	defer this.perfs.Increment(PERF_PROTO_RCV_TOTAL)

	if !msg.isValid() {
		this.perfs.Increment(PERF_PROTO_ERR_RCV_CHECKSUM)
		this.onError(errBadChecksum)
		return
	}

	msgCon := msg.Connection()
	if msgCon == nil {
		this.perfs.Increment(PERF_PROTO_ERR_RCV_CON_NIL)
		this.onError(errConnectionNil)
		return
	}

	access := this.getAccess(msgCon)
	if access < 1 {
		this.perfs.Increment(PERF_PROTO_ERR_NO_ACCESS)
		this.onError(errNoAccess)
		return
	}

	msgHeader := msg.GetHeader()
	sig       := GetMsgSig(msgHeader)

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	proc := this.sigMap[sig]

	if proc == nil {
		this.perfs.Increment(PERF_PROTO_ERR_NO_PROVIDER)
		this.onError(errors.New(fmt.Sprintf(
			"No valid message processor (sig %v). Dropping message", 
			sig,
		)))
		go msgCon.Close()
		return
	}

	if GetMsgEncryptedFlag(msgHeader) {
		if this.crypto == nil {
			this.perfs.Increment(PERF_PROTO_ERR_NO_PROVIDER)
			this.onError(errors.New(fmt.Sprintf(
				"Encryption flag set, but no encrpytion provider." +
				"Dropping message (proto: %s)",
				this.name,
			)))
			return
		}

		err := this.crypto.Decrypt(msg)
		if err != nil {
			this.perfs.Increment(PERF_PROTO_ERR_RCV_DECRYPT)
			this.onError(errors.New(fmt.Sprintf(
				"Error decrypting message (proto: %s, err: %v)",
				this.name,
				err,
			)))
			return
		}
	}

	if GetMsgCompressedFlag(msgHeader) {
		if this.compressor == nil {
			this.perfs.Increment(PERF_PROTO_ERR_NO_PROVIDER)
			this.onError(errors.New(fmt.Sprintf(
				"Compression flag set, but no compression provider."+
				"Dropping message (proto: %s)",
				this.name,
			)))
			return
		}

		err := this.compressor.Decompress(msg)
		if err != nil {
			this.perfs.Increment(PERF_PROTO_ERR_RCV_DECOMPRESS)
			this.onError(errors.New(fmt.Sprintf(
				"Error decompressing message (proto: %s, err: %v)",
				this.name,
				err,
			)))
			return
		}
	}

	dataLen  := int64(msg.Len())
	obj, err := proc.DeserializeMsg(msg, access)
	if err != nil {
		this.perfs.Increment(PERF_PROTO_ERR_DESERIALIZE)
		this.onError(errors.New(fmt.Sprintf(
			"Error deserializing message (proto: %s, err: %v)",
			this.name,
			err,
		)))
	}

	this.evtHandler.OnReceive(obj)

	this.perfs.Add(PERF_PROTO_RCV_BYTES, dataLen)
}

// sendMsg distributes the given msg to a registerd client with that id,
// if one exists.
func (this *Protocol) sendMsg(id uint32, sig uint16, obj interface{}) error {
	defer this.perfs.Increment(PERF_PROTO_SEND_TOTAL)

	this.cliMutex.RLock()
	cli := this.cliMap[id]
	this.cliMutex.RUnlock()

	if cli == nil {
		this.perfs.Increment(PERF_PROTO_ERR_SEND_INVALID_CLI)
		err := errors.New(fmt.Sprintf(
			"sendMsg failed: Client %v doesn't exist.",
			id,
		))

		this.onError(err)

		return err
	}

	this.objMutex.RLock()
	defer this.objMutex.RUnlock()

	proc := this.sigMap[sig]

	if proc == nil {
		this.perfs.Increment(PERF_PROTO_ERR_SEND_INVALID_MSG_TYPE)
		err := errors.New(fmt.Sprintf(
			"Can't send a message for an unregistered message type " +
			"signature (%v)",
			sig,
		))

		this.onError(err)

		return err
	}

	msg, err := proc.SerializeMsg(obj)
	if err != nil {
		this.perfs.Increment(PERF_PROTO_ERR_SERIALIZE)
		err := errors.New(fmt.Sprintf(
			"Serialization failure, type %d", 
			sig,
		))
		
		this.onError(err)

		return err
	}

	if GetMsgCompressedFlag(msg.GetHeader()) {
		if this.compressor == nil {
			this.perfs.Increment(PERF_PROTO_ERR_NO_PROVIDER)
			this.onError(errNoCompProvider)
			return errNoCompProvider
		}

		err := this.compressor.Compress(msg)
		if err != nil {
			this.perfs.Increment(PERF_PROTO_ERR_SEND_COMPRESS)
			this.onError(errors.New(fmt.Sprintf(
				"Error compressing data: %v", err,
			)))
			return err
		}
	}

	if GetMsgEncryptedFlag(msg.GetHeader()) {
		if this.crypto == nil {
			this.perfs.Increment(PERF_PROTO_ERR_NO_PROVIDER)
			this.onError(errNoCryptoProvider)
			return errNoCryptoProvider
		}

		err := this.crypto.Encrypt(msg)
		if err != nil {
			this.perfs.Increment(PERF_PROTO_ERR_SEND_ENCRYPT)
			this.onError(errors.New(fmt.Sprintf(
				"Error encrypting data: %v", err,
			)))
			return err
		}
	}

	timeoutSec := math.IClamp(
		msg.TimeoutSec(), 
		MIN_TIMEOUT_SEC, 
		MAX_TIMEOUT_SEC,
	)

	dataLen := int64(msg.Len())

	cli.Send(msg.GetBytes(), timeoutSec)

	this.perfs.Increment(PERF_PROTO_SEND_OK)
	this.perfs.Add(PERF_PROTO_SEND_BYTES, dataLen)

	return nil
}
