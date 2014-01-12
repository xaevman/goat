//  ---------------------------------------------------------------------------
//
//  tcp.go
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
	"github.com/xaevman/goat/lib/lifecycle"
)

// Stdlib imports.
import (
	"net"
	"sync"
	"sync/atomic"
)

// TCPSrv represents a TCP server object. The server object handles basic
// communications, client synchronization, and error handling. Client code
// only establishes a server listener via the server object. Sending and receiving
// messages is done via Protocol objects registered with the net service.
type TCPSrv struct {
	acceptChan chan *tcpCli
	discoChan  chan *tcpCli
	cliMap     map[uint32]*tcpCli
	id         uint32
	listener   net.Listener
	mutex      sync.RWMutex
	syncObj    *lifecycle.Lifecycle
}

// NewTCPSrv is a helper function which initializes a new TCPSrv instance
// and returns a pointer to it for use.
func NewTCPSrv() *TCPSrv {
	srv := TCPSrv {
		acceptChan: make(chan *tcpCli, 1),
		discoChan:  make(chan *tcpCli, 1),
		cliMap:     make(map[uint32]*tcpCli, 0),
		id:         atomic.AddUint32(&netId, 1),
		syncObj:    lifecycle.New(),
	}

	return &srv
}

// Start initializes and starts the TCP server in a new goroutine, 
// on the given network address.
func (this *TCPSrv) Start(addr string) {
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Error("%v", err)
		}

		log.Info("Startup complete %v", addr)

		this.listener = ln
		go this.acceptConnections()

		for this.syncObj.QueryRun() {
			select {
			case newCli   := <-this.acceptChan:
				this.handleConnect(newCli)
			case discoCli := <-this.discoChan:
				this.handleDisco(discoCli)
			case <-this.syncObj.QueryShutdown():
				log.Info("Shutting down %v", addr)
			}
		}

		this.syncObj.ShutdownComplete()
	}()
}

// Stop shuts the TCP server down.
func (this *TCPSrv) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, cli := range this.cliMap {
		cli.close()
	}

	this.syncObj.Shutdown()
}

// acceptConnections handles accepting connections from new clients.
func (this *TCPSrv) acceptConnections() {
	for {
		cliCon, err := this.listener.Accept()
		if err != nil {
			log.Error("%v", err)
			continue
		}

		cli := tcpCli{
			con:       cliCon.(*net.TCPConn),
			id:        atomic.AddUint32(&netId, 1),
			readChan:  make(chan []byte, 1),
			srv:       this,
			srvDisco:  this.discoChan,
			syncObj:   lifecycle.New(),
			writeChan: make(chan []byte, 1),
		}

		this.acceptChan <- &cli
	}
}

// handleConnect handles a newly connected client, registering it as a client
// of the server and startings its client handlers.
func (this *TCPSrv) handleConnect(cli *tcpCli) {
	this.mutex.Lock()
	this.cliMap[cli.id] = cli
	this.mutex.Unlock()

	log.Debug(
		"%v->%v connected", 
		cli.con.LocalAddr(),
		cli.con.RemoteAddr(),
	)

	cli.startHandlers()
}

// handleDisco handles the disconnection of a client, unregistering it from
// the list of this server's available client connections.
func (this *TCPSrv) handleDisco(cli *tcpCli) {
	this.mutex.Lock()
	delete(this.cliMap, cli.id)
	this.mutex.Unlock()

	log.Debug(
		"%v->%v disconnected",
		cli.con.LocalAddr(),
		cli.con.RemoteAddr(),
	)
}


// tcpCli represents a TCP client connection.
type tcpCli struct {
	id        uint32
	con       *net.TCPConn
	nextMsg   *NetMsg
	readChan  chan []byte
	srv       *TCPSrv
	srvDisco  chan *tcpCli
	syncObj   *lifecycle.Lifecycle
	writeChan chan []byte
}

// buildMsg is called when raw data is received off of the line. This function
// handles the segmentation of messages across multiple receive buffers or
// the packing of multiple messages into a single buffer in the stream.
func (this *tcpCli) buildMsg(msgData []byte) []byte {
	if len(msgData) < 1 {
		return nil
	}

	if this.nextMsg == nil {
		// search for a good header
		if !ValidateMsgHeader(msgData) {
			if len(msgData) > 7 {
				return msgData[1:]
			}
		}

		// start msg object for this message
		size        := GetMsgSize(msgData)
		this.nextMsg = &NetMsg {
			cli:    this,
			data:   make([]byte, size),
			header: GetMsgHeader(msgData),
		}
	}

	// read available data, up to msg length
	leftovers, complete := this.nextMsg.addData(msgData[4:])
	if complete {
		// dispatch message
		routeMsg(this.nextMsg)
		this.nextMsg = nil
	}

	if leftovers != nil {
		return leftovers
	}

	return nil
}

// close shuts down the client TCP connection.
func (this *tcpCli) close() {
	this.con.Close()
}

// handleReads runs in its own goroutine, looping endlessly, reading data
// of of the line.
func (this *tcpCli) handleReads() {
	var count int
	var err   error
	
	buffer := make([]byte, 1024)

	for {
		count, err = this.con.Read(buffer)
		// disco
		if count < 1 {
			this.notifyDisco()
			return
		}

		// received data
		this.readChan<- buffer[:count]

		// a real error
		if err != nil {
			log.Error("%v", err)
			continue
		}
	}
}

// handleWrites runs in its own goroutine, looping endlessly, putting
// write events onto the line.
func (this *tcpCli) handleWrites() {
	var count int
	var err   error

	for {
		select {
		case data := <-this.writeChan:
			count, err = this.con.Write(data)
			// disco
			if count < 1 {
				this.notifyDisco()
				return
			}

			if err != nil {
				log.Error("%v", err)
				continue
			}
		}
	}
}

// runCli runs in its own goroutine, handling read events that bubble up
// from the handleReads goroutine, building message objects, and sending them
// up the pipeline to registered protocols. runCli is also responsible for 
// handling and coordinating shutdown of a tcpCli when signaled.
func (this *tcpCli) runCli() {
	for this.syncObj.QueryRun() {
		select {
		case buffer := <-this.readChan:
			extra := this.buildMsg(buffer)
			for extra != nil {
				extra = this.buildMsg(extra)
			}
		case <-this.syncObj.QueryShutdown():
		}
	}

	this.close()

	this.syncObj.ShutdownComplete()
}

// notifyDisco bubbles a disco event up to the tcpCli's parent TCPSrv so 
// that it can properly handle the client's disconnection.
func (this *tcpCli) notifyDisco() {
	this.syncObj.Shutdown()

	if this.srvDisco != nil {
		this.srvDisco<- this
	}
}

// startHandlers starts the 3 goroutines responsible for handling IO and 
// synchronization for this client.
func (this *tcpCli) startHandlers() {
	go this.runCli()
	go this.handleReads()
	go this.handleWrites()
}

// writeData accepts write events and forwards them to the handleWrites goroutine.
func (this *tcpCli) writeData(data []byte) {
	this.writeChan<- data
}
