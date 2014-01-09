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
		acceptChan: make(chan *tcpCli, 1),
		discoChan:  make(chan *tcpCli, 1),
		cliMap:     make(map[uint32]*tcpCli, 0),
		id:         atomic.AddUint32(&netId, 1),
		syncObj:    lifecycle.New(),
	}

	return &srv
}


type TCPSrv struct {
	acceptChan chan *tcpCli
	discoChan  chan *tcpCli
	cliMap     map[uint32]*tcpCli
	id         uint32
	listener   net.Listener
	mutex      sync.RWMutex
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

		cli := tcpCli{
			con:       cliCon.(*net.TCPConn),
			id:        atomic.AddUint32(&netId, 1),
			readChan:  make(chan []byte, 1),
			srvDisco:  this.discoChan,
			syncObj:   lifecycle.New(),
			writeChan: make(chan []byte, 1),
		}

		this.acceptChan <- &cli
	}
}

func (this *TCPSrv) handleConnect(cli *tcpCli) {
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

func (this *TCPSrv) handleDisco(cli *tcpCli) {
	this.mutex.Lock()
	delete(this.cliMap, cli.id)
	this.mutex.Unlock()

	log.Info(
		"%v->%v disconnected",
		cli.con.LocalAddr(),
		cli.con.RemoteAddr(),
	)
}


type tcpCli struct {
	id        uint32
	con       *net.TCPConn
	readChan  chan []byte
	srvDisco  chan *tcpCli
	syncObj   *lifecycle.Lifecycle
	writeChan chan []byte
}

func (this *tcpCli) close() {
	this.con.Close()
}

func (this *tcpCli) handleReads() {
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

func (this *tcpCli) handleWrites() {
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

func (this *tcpCli) runCli() {
	for this.syncObj.QueryRun() {
		select {
		case <-this.readChan:
		case <-this.syncObj.QueryShutdown():
		}
	}

	this.close()

	this.syncObj.ShutdownComplete()
}

func (this *tcpCli) sendDisco() {
	this.syncObj.Shutdown()

	if this.srvDisco != nil {
		this.srvDisco<- this
	}
}

func (this *tcpCli) startHandlers() {
	go this.runCli()
	go this.handleReads()
	go this.handleWrites()
}

func (this *tcpCli) writeData(data []byte) {
	this.writeChan<- data
}
