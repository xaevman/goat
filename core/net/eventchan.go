//  ---------------------------------------------------------------------------
//
//  eventchan.go
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
)

// Stdlib imports.
import (
	"time"
)


// EventChan implements the EventHandler interface and exposes a
// channel based way of receiving events from underlying net objects.
type EventChan struct {
	conChan     chan Connection
	discoChan   chan Connection
	timeoutChan chan *TimeoutEvent
}

// NewEventChan is a helper construction function which creates and
// initializes a new EventChan object, sets it as the EventHandler for
// the net service, and returns a pointer to it for direct use.
func NewEventChan() *EventChan {
	ec := EventChan{
		conChan     : make(chan Connection, 0),
		discoChan   : make(chan Connection, 0),
		timeoutChan : make(chan *TimeoutEvent, 0),
	}

	SetEventHandler(&ec)

	return &ec
}

// OnConnect is called by the net service when a new socket connects.
// This should not be used directly by user code. Instead you should
// insert QueryConnect() into your channel based IO logic, where
// you'll receive signals containing the Connection objects representing
// sockets which have just connected.
func (this *EventChan) OnConnect(con Connection) {
	select {
	case this.conChan <- con:
	case <-time.After(DEFAULT_TIMEOUT_SEC * time.Second):
		onTimeout(TIMEOUT_CONNECT, 0, con.Id(), con)
	}
}

// OnDisconnect is called by the net service when a socket disconnects.
// This should not be used directly by user code. Instead you should
// insert QueryDisconnect() into your channel based IO logic, where
// you'll receive signals containing the Connection objects representin
// sockets which have been disconnected.
func (this *EventChan) OnDisconnect(con Connection) {
	select {
	case this.discoChan <- con:
	case <-time.After(DEFAULT_TIMEOUT_SEC * time.Second):
		onTimeout(TIMEOUT_DISCONNECT, 0, con.Id(), con)
	}
}

// OnTimeout is called by the net service when a timeout error bubbles
// up from any lower level network providers. This function should not be
// called directly by user code. Instead, you should report your own timeout
// errors by calling net.Timeout(), and you should query for timeouts by
// inserting QueryTimeout() into your channel based IO logic.
func (this *EventChan) OnTimeout(timeout *TimeoutEvent) {
	select {
	case this.timeoutChan <- timeout:
	case <-time.After(DEFAULT_TIMEOUT_SEC * time.Second):
		log.Error("Timeout sending timeout event (ouch): %v", timeout)
	}
}

// QueryConnect exposes a read-only channel which signals the connection
// of a new Connection.
func (this *EventChan) QueryConnect() <-chan Connection {
	return this.conChan
}

// QueryDisconnect exposes a read-only channel which signals the disconnection
// of an underlying Connection.
func (this *EventChan) QueryDisconnect() <-chan Connection {
	return this.discoChan
}

// QueryTimeout exposes a read-only channel which signals timeout events
// from underlying network services.
func (this *EventChan) QueryTimeout() <-chan *TimeoutEvent {
	return this.timeoutChan
}
