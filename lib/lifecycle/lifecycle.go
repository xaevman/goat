//  ---------------------------------------------------------------------------
//
//  lifecycle.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package lifecycle

// Stdlib imports.
import (
	"sync"
	"time"
)

// Lifecycle provides a simple way to coordinate the lifecycle of
// worker goroutines that loop infinitely
type Lifecycle struct {
	run           bool
	heartbeatChan chan bool
	heartbeatDur  time.Duration
	lock          sync.RWMutex
	shutdownChan  chan bool
	waitChan      chan bool
}

// New is a helper function which creates a newly initialized Lifecycle
// object and returns a pointer to it for use.
func New() *Lifecycle {
	newObj := Lifecycle {
		run: 		   true,
		heartbeatChan: make(chan bool, 1),
		heartbeatDur:  time.Duration(0),
		shutdownChan:  make(chan bool),
		waitChan:      make(chan bool),
	}

	return &newObj
}

// QueryHeartbeat returns the read-only channel on which a user should
// listen for a periodic heartbeat signal.
func (this *Lifecycle) QueryHeartbeat() <-chan bool {
	return this.heartbeatChan
}

// QueryRun returns the current status of the Sync object. It will return
// true until Shutdown() is called. Client code can use this function as
// their criteria in a typical for-select work loop.
func (this *Lifecycle) QueryRun() bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.run
}

// QueryShutdown returns the read-only channel on which a user should
// listen for a signal to shut down.
func (this *Lifecycle) QueryShutdown() <-chan bool {
	return this.shutdownChan
}

// Shutdown sets the run flag to false, sends the shutdown signal back
// to the client on the shutdown channel, and then blocks until the client
// calls ShutdownComplete()
func (this *Lifecycle) Shutdown() {
	this.lock.Lock()

	if !this.run {
		this.lock.Unlock()
		return
	}

	this.StopHeart()
	this.run = false
	
	this.lock.Unlock()

	close(this.shutdownChan)
	<-this.waitChan
}

// ShutdownComplete sends a signal which unblocks the call to Shutdown(). Client
// code should call this function once its shutdown procedures are completed.
func (this *Lifecycle) ShutdownComplete() {
	close(this.waitChan)
}

// StartHeart sets the heartbeat cycle time in milliseconds and starts
// the heartbeat timer.
func (this *Lifecycle) StartHeart(heartbeatMs int) {
	this.heartbeatDur = time.Duration(heartbeatMs) * time.Millisecond
	go this.heartbeat()
}

// StopHeart sets the heartbeat cyle time to zero. If any timers are outstanding
// they will stil fire, but a signal will not be sent back to clients.
func (this *Lifecycle) StopHeart() {
	this.heartbeatDur = time.Duration(0)
}

// heartbeat sends the heartbeat signal back to the client on the heartbeat 
// channel and schedules the next heartbeat for heartbeatMs milliseconds in 
// the future. If heartbeatMs is less than 1, the heartbeat is disabled.
func (this *Lifecycle) heartbeat() {
	this.lock.RLock()
	if !this.run || this.heartbeatDur == 0 {
		this.lock.RUnlock()
		return
	}
	this.lock.RUnlock()

	this.heartbeatChan<- true

	time.AfterFunc(
        this.heartbeatDur,
        this.heartbeat,
    )
}
