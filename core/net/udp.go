//  ---------------------------------------------------------------------------
//
//  udp.go
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
	"errors"
	"fmt"
	stdnet "net"
	"sync"
	"time"
)

// Perf counters.
const (
	PERF_UDP_MSG_RECEIVE = iota
	PERF_UDP_MSG_RECEIVE_BYTES
	PERF_UDP_MSG_SEND
	PERF_UDP_MSG_SEND_BYTES
	PERF_UDP_MSG_TIMEOUT
	PERF_UDP_SERVERS
	PERF_UDP_COUNT
)

// Perf counter friendly names.
var udpPerfNames = []string {
	"MsgReceived",
	"MsgReceivedBytes",
	"MsgSent",
	"MsgSentBytes",
	"MsgTimeout",
	"Servers",
}

// Global tcp perf object.
var udpPerfs = perf.NewCounterSet(
	"Module.Net.Udp",
	PERF_UDP_COUNT,
	udpPerfNames,
)


// newudpSrv is a constructor function which initializes a new udpSrv
// instance and returns a pointer to it for use.
func newudpSrv(proto *Protocol) *udpSrv {
	newsrv      := new(udpSrv)
	newsrv.proto = proto

	return newsrv
}

// udpSrv represents a UDP endpoint which actively listens for new
// packets coming into its address. The server object handles basic
// communications, client synchronization, and error handling.
type udpSrv struct {
	addr    stdnet.Addr
	closing bool
	mutex   sync.RWMutex
	proto   *Protocol
	socket  *stdnet.UDPConn
}

// Start initializes and starts the UDP server on the given network address.
func (this *udpSrv) Start(addr string) (Connection, error) {
	addrObj, err := stdnet.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	this.addr = addrObj

	con, err := stdnet.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	udpCon, ok := con.(*stdnet.UDPConn)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Invalid type %T", con))
	}

	this.socket = udpCon
	
	this.startHandlers()

	log.Info("UDP start complete %v", addr)

	udpPerfs.Increment(PERF_UDP_SERVERS)

	return nil, nil
}

// Stop marks the UDP server for shutdown and closes the socket.
func (this *udpSrv) Stop() {
	this.mutex.Lock()
	this.closing = true
	this.mutex.Unlock()

	this.socket.Close()

	udpPerfs.Add(PERF_UDP_SERVERS, -1)
}

// handleReads accepts new packets coming into the UDP server, formats
// them into net.Msg objects and bubbles them up to the protocol layer
// for handling.
func (this *udpSrv) handleReads() {
	inbuffer := make([]byte, MAX_NET_MSG_LEN)

	for {
		count, addr, err := this.socket.ReadFrom(inbuffer)
		if count < 1 {
			if this.isClosing() {
				return 
			}

			continue
		}

		if err != nil {
			if this.isClosing() {
				return
			}

			log.Error(err.Error())
			continue
		}

		if count > MAX_NET_MSG_LEN {
			log.Error(
				"Message > MAX_NET_MSG_LEN (%d > %d)", 
				count, 
				MAX_NET_MSG_LEN,
			)
		}

		con, err := this.proto.getUDPEndpoint(addr, this.socket)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		cpBuffer := make([]byte, count)
		copy(cpBuffer, inbuffer[:count])

		msg := NewMsg()
		msg.SetConnection(con)

		_, complete := msg.addData(cpBuffer)
		if !complete {
			log.Error("Received incomplete datagram. Dropping...")
		}

		this.notifyMsg(msg)
	}
}

// isClosing is a helper function which checks to see if the UDP server
// has been marked for shutdown.
func (this *udpSrv) isClosing() bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	
	return this.closing
}

// notifyMsg bubbles a received msg up to the parent object.
func (this *udpSrv) notifyMsg(msg *Msg) {
	udpPerfs.Increment(PERF_UDP_MSG_RECEIVE)
	udpPerfs.Add(PERF_UDP_MSG_RECEIVE_BYTES, int64(msg.Len()))

	select {
	case this.proto.rcvChan<- msg:
	case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		log.Error("notifyMsg timeout")
	}	
}

// startHandlers starts the go routine responsible for handling reads
// coming into this server endpoint.
func (this *udpSrv) startHandlers() {
	go this.handleReads()
}


// udpEndpoint represents a UDP endpoint to which you will send
// data.
type udpEndpoint struct {
	discoChan   chan Connection
	id          uint32
	key         string
	remoteAddr  stdnet.Addr
	socket      *stdnet.UDPConn
	syncObj     *lifecycle.Lifecycle
	timeoutChan chan *TimeoutEvent
	writeChan   chan []byte
}

// close shuts down the client UDP endpoint.
func (this *udpEndpoint) Close() {
	this.socket.Close()
}

// Id returns the net service id of this udp endpoint.
func (this *udpEndpoint) Id() uint32 {
	return this.id
}

// Key returns the key information assigned to this udp endpoint.
func (this *udpEndpoint) Key() string {
	return this.key
}

// LocalAddr returns the local endpoint address.
func (this *udpEndpoint) LocalAddr() stdnet.Addr {
	return this.socket.LocalAddr()
}

// RemoteAddr returns the remote endpoint's address.
func (this *udpEndpoint) RemoteAddr() stdnet.Addr {
	if this.remoteAddr == nil {
		return this.socket.RemoteAddr()
	}

	return this.remoteAddr
}

// Send takes raw data and sends it to the endpoint's write go routine.
func (this *udpEndpoint) Send(data []byte, timeoutSec int) {
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

// handleWrites runs in its own goroutine, looping endlessly, putting
// write events onto the line.
func (this *udpEndpoint) handleWrites() {
	var count int
	var err   error

	for this.syncObj.QueryRun() {
		select {
		case data := <-this.writeChan:
			if this.remoteAddr == nil {
				count, err = this.socket.Write(data)
			} else {
				count, err = this.socket.WriteTo(data, this.remoteAddr)
			}

			if err != nil {
				log.Error(err.Error())
				continue
			}

			if count < 1 {
				this.notifyDisco()
				return
			}

			udpPerfs.Increment(PERF_UDP_MSG_SEND)
			udpPerfs.Add(PERF_UDP_MSG_SEND_BYTES, int64(count))

		case <-time.After(QUEUE_TIMEOUT_SEC * time.Second):
		case <-this.syncObj.QueryShutdown():
		}
	}

	this.syncObj.ShutdownComplete()
}

// notifyDisco bubbles a disco event up to the udpCon's parent object.
func (this *udpEndpoint) notifyDisco() {
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

// notifyTimeout bubbles a timeout event up to the udpCon's parent object.
func (this *udpEndpoint) notifyTimeout(
	kind int, 
	sig  uint16, 
	id   uint32, 
	data interface{},
) {
	udpPerfs.Increment(PERF_UDP_MSG_TIMEOUT)

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

// startHandlers starts the  goroutine responsible for handling IO and
// for this endpoint.
func (this *udpEndpoint) startHandlers() {
	go this.handleWrites()
}
