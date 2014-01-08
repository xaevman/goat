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

// NewTCPSrv is a helper function which initializes a new TCPSrv instance
// and returns a pointer to it for use.
func NewTCPSrv() *TCPSrv {
	srv := TCPSrv {
		acceptChan: make(chan *remoteCli, 1),
		discoChan:  make(chan *remoteCli, 1),
		cliMap:     make(map[uint64]*remoteCli, 0),
		id:         atomic.AddUint64(&netId, 1),
		readChan:   make(chan *RawMsg, 1),
		syncObj:    lifecycle.New(),
	}

	return &srv
}


type TCPSrv struct {
	acceptChan chan *remoteCli
	discoChan  chan *remoteCli
	cliMap     map[uint64]*remoteCli
	id         uint64
	listener   net.Listener
	mutex      sync.Mutex
	readChan   chan *RawMsg
	syncObj    *lifecycle.Lifecycle
}

func (this *TCPSrv) Start(addr string) {
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
		case newMsg   := <-this.readChan:
			log.Info(
				"[cli %v] msg len %v", 
				newMsg.cli.id, 
				len(newMsg.data),
			)
		case <-this.syncObj.QueryShutdown():
			log.Info("Shutting down %v", addr)
		}
	}

	this.syncObj.ShutdownComplete()
}

func (this *TCPSrv) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, cli := range this.cliMap {
		cli.close()
	}

	this.syncObj.Shutdown()
}

func (this *TCPSrv) acceptConnections() {
	for {
		cliCon, err := this.listener.Accept()
		if err != nil {
			log.Error("%v", err)
			continue
		}

		cli := remoteCli{
			con:       cliCon.(*net.TCPConn),
			id:        atomic.AddUint64(&netId, 1),
			readChan:  make(chan []byte, 1),
			srvDisco:  this.discoChan,
			srvRead:   this.readChan,
			syncObj:   lifecycle.New(),
			writeChan: make(chan []byte, 1),
		}

		this.acceptChan <- &cli
	}
}

func (this *TCPSrv) handleConnect(cli *remoteCli) {
	this.mutex.Lock()
	this.cliMap[cli.id] = cli
	this.mutex.Unlock()

	log.Info(
		"%v->%v connected", 
		cli.con.LocalAddr(),
		cli.con.RemoteAddr(),
	)

	cli.startHandlers()
}

func (this *TCPSrv) handleDisco(cli *remoteCli) {
	this.mutex.Lock()
	delete(this.cliMap, cli.id)
	this.mutex.Unlock()

	log.Info(
		"%v->%v disconnected",
		cli.con.LocalAddr(),
		cli.con.RemoteAddr(),
	)
}


type remoteCli struct {
	id        uint64
	con       *net.TCPConn
	readChan  chan []byte
	srvDisco  chan *remoteCli
	srvRead   chan *RawMsg
	syncObj   *lifecycle.Lifecycle
	writeChan chan []byte
}

func (this *remoteCli) close() {
	this.con.Close()
}

func (this *remoteCli) handleReads() {
	var count int
	var err   error
	
	buffer := make([]byte, 1024)

	for {
		count, err = this.con.Read(buffer)
		// disco
		if count < 1 {
			this.sendDisco()
			return
		}

		// a real error
		if err != nil {
			log.Error("%v", err)
			continue
		}

		// received data
		this.readChan<- buffer[:count]
	}
}

func (this *remoteCli) handleWrites() {
	var count int
	var err   error

	for {
		select {
		case data := <-this.writeChan:
			count, err = this.con.Write(data)
			// disco
			if count < 1 {
				this.sendDisco()
				return
			}

			if err != nil {
				log.Error("%v", err)
				continue
			}
		}
	}
}

func (this *remoteCli) runCli() {
	for this.syncObj.QueryRun() {
		select {
		case data := <-this.readChan:
			msg := RawMsg {
				cli:  this,
				data: data,
			}
			this.srvRead<- &msg
		case <-this.syncObj.QueryShutdown():
		}
	}

	this.close()

	this.syncObj.ShutdownComplete()
}

func (this *remoteCli) sendDisco() {
	this.syncObj.Shutdown()

	if this.srvDisco != nil {
		this.srvDisco<- this
	}
}

func (this *remoteCli) startHandlers() {
	go this.runCli()
	go this.handleReads()
	go this.handleWrites()
}

func (this *remoteCli) writeData(data []byte) {
	this.writeChan<- data
}
