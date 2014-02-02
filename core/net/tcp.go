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
)

// Stdlib imports.
import (
	"net"
	"sync/atomic"
	"time"
)

// TCP Buffer size for reading data off the line.
const TCP_BUFFER_SIZE_B = 1 * 1024 // 1KB


// newtcpCli is a helper constructor function which returns a pointer to a
// newly intialized tcpCli object for use.
func newtcpCli(proto *Protocol) *tcpCli {
	cli := tcpCli{
		id       : atomic.AddUint32(&netId, 1),
		protocol : proto,
	}

	return &cli
}

// tcpCli represents a TCP client object. The client object handles connecting
// to server processes and manages basic IO and synchronization tasks, accepting
// data from, and passing data up to, the net service's registered Protocol
// objects.
type tcpCli struct {
	id        uint32
	protocol  *Protocol
	srvSocket *tcpCon
}

// Start connects the tcpCli object to a new server endpoint.
func (this *tcpCli) Start(addr string) (Connection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}

	tCon := tcpCon{
		discoChan   : this.protocol.discoChan,
		id          : atomic.AddUint32(&netId, 1),
		rcvChan     : this.protocol.rcvChan,
		readChan    : make(chan []byte),
		socket      : conn.(*net.TCPConn),
		syncObj     : lifecycle.New(),
		timeoutChan : this.protocol.timeoutChan,
		writeChan   : make(chan []byte),
	}

	tCon.startHandlers()

	this.srvSocket = &tCon

	return &tCon, nil
}

// Stop closes the tcpCli's existing socket as well as all of the
// client's go routines.
func (this *tcpCli) Stop() {
	this.srvSocket.Close()
}


// newtcpSrv is a helper function which initializes a new tcpSrv instance
// and returns a pointer to it for use.
func newtcpSrv(proto *Protocol) *tcpSrv {
	srv := tcpSrv{
		id       : atomic.AddUint32(&netId, 1),
		protocol : proto,
		syncObj  : lifecycle.New(),
	}

	return &srv
}

// tcpSrv represents a TCP server object. The server object handles basic
// communications, client synchronization, and error handling. Client code
// only establishes a server listener via the server object. Sending and receiving
// messages is done via Protocol objects registered with the net service.
type tcpSrv struct {
	id       uint32
	listener net.Listener
	protocol *Protocol
	syncObj  *lifecycle.Lifecycle
}

// Start initializes and starts the TCP server in a new goroutine,
// on the given network address.
func (this *tcpSrv) Start(addr string) (Connection, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}

	this.listener = ln	
	log.Info("Startup complete %v", addr)

	go this.acceptConnections()

	return nil, nil
}

// Stop shuts the TCP server down.
func (this *tcpSrv) Stop() {
	go this.syncObj.Shutdown()
	this.listener.Close()
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
			id          : atomic.AddUint32(&netId, 1),
			rcvChan     : this.protocol.rcvChan,
			readChan    : make(chan []byte),
			socket      : cliCon.(*net.TCPConn),
			syncObj     : lifecycle.New(),
			timeoutChan : this.protocol.timeoutChan,
			writeChan   : make(chan []byte),
		}

		this.protocol.connectChan <- &cli

		cli.startHandlers()
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
	readChan    chan []byte
	socket      *net.TCPConn
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
func (this *tcpCon) LocalAddr() net.Addr {
	return this.socket.LocalAddr()
}

// RemoteAddr returns the remote endpoint's address for this tcp connection.
func (this *tcpCon) RemoteAddr() net.Addr {
	return this.socket.RemoteAddr()
}

// Send takes raw data and sends it to the connection's write go routine.
func (this *tcpCon) Send(data []byte, timeoutSec int) {
	select {
	case this.writeChan <- data:
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		sig := GetMsgSig(GetMsgHeader(data))
		this.notifyTimeout(TIMEOUT_SEND, sig, this.id, data)
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
	var err error

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
		this.readChan <- cpBuffer

		// a real error
		if err != nil {
			log.Error("%v", err)
			continue
		}
	}
}

// handleWrites runs in its own goroutine, looping endlessly, putting
// write events onto the line.
func (this *tcpCon) handleWrites() {
	var count int
	var err error

	for {
		select {
		case data := <-this.writeChan:
			count, err = this.socket.Write(data)
			// disco
			if count < 1 {
				this.notifyDisco()
				return
			}

			if err != nil {
				log.Error("%v", err)
			}
		}
	}
}

// notifyDisco bubbles a disco event up to the tcpCon's parent object.
func (this *tcpCon) notifyDisco() {
	this.syncObj.Shutdown()

	if this.discoChan == nil {
		return
	}

	this.discoChan<- this
}

// notifyMsg bubbles a received msg up to the tcpCon's parent object.
func (this *tcpCon) notifyMsg(msg *Msg) {
	if this.rcvChan == nil {
		return
	}

	this.rcvChan<- msg
}

// notifyTimeout bubbles a timeout event up to the tcpCon's parent object.
func (this *tcpCon) notifyTimeout(
	kind int, 
	sig  uint16, 
	id   uint32, 
	data interface{},
) {
	if this.timeoutChan == nil {
		return
	}

	timeout            := new(TimeoutEvent)
	timeout.Data        = data
	timeout.MessageType = sig
	timeout.ParentId    = id
	timeout.TimeoutType = kind

	this.timeoutChan<- timeout
}

// runCli runs in its own goroutine, handling read events that bubble up
// from the handleReads goroutine, building message objects, and sending them
// up the pipeline to registered protocols. runCli is also responsible for
// handling and coordinating shutdown of a tcpCon when signaled.
func (this *tcpCon) runCli() {
	for this.syncObj.QueryRun() {
		select {
		case data := <-this.readChan:
			for 
				pending := this.buildMsg(data)
				pending != nil
				pending = this.buildMsg(pending) {}
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
