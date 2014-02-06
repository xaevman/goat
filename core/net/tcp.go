//  ---------------------------------------------------------------------------
//
//  tcp.go
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
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	stdnet "net"
	"time"
)

// Perf counters.
const (
	PERF_TCP_CONNECTIONS = iota
	PERF_TCP_DISCO
	PERF_TCP_MSG_RECEIVE
	PERF_TCP_MSG_RECEIVE_BYTES
	PERF_TCP_MSG_SEND
	PERF_TCP_MSG_SEND_BYTES
	PERF_TCP_MSG_TIMEOUT
	PERF_TCP_SERVERS
	PERF_TCP_COUNT
)

// Perf counter friendly names.
var tcpPerfNames = []string {
	"Connections",
	"Disconnect",
	"MsgReceived",
	"MsgReceivedBytes",
	"MsgSent",
	"MsgSentBytes",
	"MsgTimeout",
	"Servers",
}

// Global tcp perf object.
var tcpPerfs = perf.NewCounterSet(
	"Module.Net.Tcp",
	PERF_TCP_COUNT,
	tcpPerfNames,
)

// TCP Buffer size for reading data off the line.
const TCP_BUFFER_SIZE_B = 256


// newtcpSrv is a helper function which initializes a new tcpSrv instance
// and returns a pointer to it for use.
func newtcpSrv(proto *Protocol) *tcpSrv {
	srv := tcpSrv{
		protocol : proto,
		syncObj  : lifecycle.New(),
	}

	return &srv
}

// tcpSrv represents a TCP server object. The server object handles basic
// communications, client synchronization, and error handling.
type tcpSrv struct {
	listener stdnet.Listener
	protocol *Protocol
	syncObj  *lifecycle.Lifecycle
}

// Start initializes and starts the TCP server in a new goroutine,
// on the given network address.
func (this *tcpSrv) Start(addr string) (Connection, error) {
	ln, err := stdnet.Listen("tcp", addr)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}

	this.listener = ln	
	log.Info("TCP start complete %v", addr)

	go this.acceptConnections()

	tcpPerfs.Increment(PERF_TCP_SERVERS)

	return nil, nil
}

// Stop shuts the TCP server down.
func (this *tcpSrv) Stop() {
	go this.syncObj.Shutdown()
	this.listener.Close()

	tcpPerfs.Add(PERF_TCP_SERVERS, -1)
}

// acceptConnections handles accepting connections from new clients.
func (this *tcpSrv) acceptConnections() {
	for {
		cliCon, err := this.listener.Accept()
		if err != nil {
			if !this.syncObj.QueryRun() { 
				break 
			}

			log.Error(err.Error())
			continue
		}

		cli := tcpCon{
			discoChan   : this.protocol.discoChan,
			id          : NextNetID(),
			rcvChan     : this.protocol.rcvChan,
			socket      : cliCon.(*stdnet.TCPConn),
			syncObj     : lifecycle.New(),
			timeoutChan : this.protocol.timeoutChan,
			writeChan   : make(chan []byte, QUEUE_BUFFERS),
		}

		this.protocol.connectChan <- &cli

		cli.startHandlers()

		tcpPerfs.Increment(PERF_TCP_CONNECTIONS)
	}

	this.syncObj.ShutdownComplete()
}


// tcpCon represents a TCP connection.
type tcpCon struct {
	discoChan   chan Connection
	id          uint32
	key         string
	nextMsg     *Msg
	rcvChan     chan *Msg
	socket      *stdnet.TCPConn
	syncObj     *lifecycle.Lifecycle
	timeoutChan chan *TimeoutEvent
	writeChan   chan []byte
}

// close shuts down the client TCP connection.
func (this *tcpCon) Close() {
	this.socket.Close()
}

// Id returns the net service id of this tcp connection.
func (this *tcpCon) Id() uint32 {
	return this.id
}

// Key returns the key information assigned to this tcp connection.
func (this *tcpCon) Key() string {
	return this.key
}

// LocalAddr returns the local endpoint address for this tcp connection.
func (this *tcpCon) LocalAddr() stdnet.Addr {
	return this.socket.LocalAddr()
}

// RemoteAddr returns the remote endpoint's address for this tcp connection.
func (this *tcpCon) RemoteAddr() stdnet.Addr {
	return this.socket.RemoteAddr()
}

// Send takes raw data and sends it to the connection's write go routine.
func (this *tcpCon) Send(data []byte, timeoutSec int) {
	select {
	case this.writeChan <- data:
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		sig         := uint16(0)
		header, err := GetMsgHeader(data)
		if err == nil {
			sig = GetMsgSig(header)
		}
		this.notifyTimeout(TIMEOUT_SEND, sig, this.id, data)
	case <-this.syncObj.QueryShutdown():
		return
	}
}

// buildMsg is called when raw data is received off of the line. This function
// handles the segmentation of messages across multiple receive buffers or
// the packing of multiple messages into a single buffer in the stream. buildMsg
// returns any extra buffer data leftover after processing.
func (this *tcpCon) buildMsg(msgData []byte) []byte {
	if len(msgData) < 1 {
		return nil
	}

	if this.nextMsg == nil {
		this.nextMsg = NewMsg()
		this.nextMsg.SetConnection(this)
	}

	// read available data, up to msg length
	leftovers, complete := this.nextMsg.addData(msgData)
	if complete {
		// dispatch message
		dispatchMsg := this.nextMsg
		this.nextMsg = nil
		this.notifyMsg(dispatchMsg)
	}

	return leftovers
}

// handleReads runs in its own goroutine, looping endlessly, reading data
// of of the line.
func (this *tcpCon) handleReads() {
	var count int
	var err   error

	buffer := make([]byte, TCP_BUFFER_SIZE_B)

	for {
		count, err = this.socket.Read(buffer)
		// disco
		if count < 1 {
			this.notifyDisco()
			return
		}

		// received data
		cpBuffer := make([]byte, count)
		copy(cpBuffer, buffer[:count])

		for 
			pending := this.buildMsg(cpBuffer)
			pending != nil
			pending = this.buildMsg(pending) {}

		// a real error
		if err != nil {
			log.Error("%v", err)
		}
	}
}

// handleWrites runs in its own goroutine, looping endlessly, putting
// write events onto the line.
func (this *tcpCon) handleWrites() {
	var count int
	var err   error

	for this.syncObj.QueryRun() {
		select {
		case data := <-this.writeChan:
			count, err = this.socket.Write(data)
			tcpPerfs.Increment(PERF_TCP_MSG_SEND)
			tcpPerfs.Add(PERF_TCP_MSG_SEND_BYTES, int64(len(data)))

			// disco
			if count < 1 {
				this.notifyDisco()
				return
			}

			if err != nil {
				log.Error("%v", err)
			}
		case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		case <-this.syncObj.QueryShutdown():
		}
	}
}

// notifyDisco bubbles a disco event up to the tcpCon's parent object.
func (this *tcpCon) notifyDisco() {
	tcpPerfs.Add(PERF_TCP_CONNECTIONS, -1)

	this.syncObj.Shutdown()

	if this.discoChan == nil {
		return
	}

	select {
	case this.discoChan<- this:
	case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		log.Error("notifyDisco timeout")
	case <-this.syncObj.QueryShutdown():
		return
	}
}

// notifyMsg bubbles a received msg up to the tcpCon's parent object.
func (this *tcpCon) notifyMsg(msg *Msg) {
	tcpPerfs.Increment(PERF_TCP_MSG_RECEIVE)
	tcpPerfs.Add(PERF_TCP_MSG_RECEIVE_BYTES, int64(msg.Len()))

	if this.rcvChan == nil {
		return
	}

	select {
	case this.rcvChan<- msg:
	case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		log.Error("notifyMsg timeout")
	case <-this.syncObj.QueryShutdown():
		return
	}	
}

// notifyTimeout bubbles a timeout event up to the tcpCon's parent object.
func (this *tcpCon) notifyTimeout(
	kind int, 
	sig  uint16, 
	id   uint32, 
	data interface{},
) {
	tcpPerfs.Increment(PERF_TCP_MSG_TIMEOUT)

	if this.timeoutChan == nil {
		return
	}

	timeout            := new(TimeoutEvent)
	timeout.Data        = data
	timeout.MessageType = sig
	timeout.ParentId    = id
	timeout.TimeoutType = kind

	select {
	case this.timeoutChan<- timeout:
	case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		log.Error("notifyTimeout timeout")
	case <-this.syncObj.QueryShutdown():
		return
	}
}

// runCli runs in its own goroutine, handling read events that bubble up
// from the handleReads goroutine, building message objects, and sending them
// up the pipeline to registered protocols. runCli is also responsible for
// handling and coordinating shutdown of a tcpCon when signaled.
func (this *tcpCon) runCli() {
	for this.syncObj.QueryRun() {
		select {
		case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		case <-this.syncObj.QueryShutdown():
		}
	}

	this.Close()

	this.syncObj.ShutdownComplete()
}

// startHandlers starts the 3 goroutines responsible for handling IO and
// synchronization for this client.
func (this *tcpCon) startHandlers() {
	go this.runCli()
	go this.handleReads()
	go this.handleWrites()
}
