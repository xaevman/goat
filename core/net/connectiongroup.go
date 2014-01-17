//  ---------------------------------------------------------------------------
//
//  connectiongroup.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// Stdlib imports.
import (
	"net"
	"sync"
	"sync/atomic"
)

// Send distribution schemes.
const (
	BROADCAST = iota
	ROUND_ROBIN
)

// ConnectionGroup represents a group of Connections, and itself implements 
// the Connection interface so that it can be used interchangebly with single
// Connection objects.
type ConnectionGroup struct {
	conList     []Connection
	id          uint32
	key         string
	lastIndex   int
	mutex       sync.RWMutex
	name        string
	routeScheme int
}

// NewConnectionGroup is a constructor helper which builds a newly initalized 
// instance of ConnectionGroup and returns a pointer to it for use.
func NewConnectionGroup(name string, scheme int) *ConnectionGroup {
	conGroup := ConnectionGroup {
		conList:     make([]Connection, 0),
		id:          atomic.AddUint32(&netId, 1),
		name:        name,
		routeScheme: scheme,
	}

	return &conGroup
}

// AddConnection adds a new Connection object to this group.
func (this *ConnectionGroup) AddConnection(con Connection) {
	if con == nil {
		return
	}

	this.mutex.RLock()
	for i := range this.conList {
		if this.conList[i] == con {
			this.mutex.RUnlock()
			return
		}
	}
	this.mutex.RUnlock()

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.conList = append(this.conList, con)
}

// Close removes all Connection objects from the group, after calling Close()
// on each.
func (this *ConnectionGroup) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for i := range this.conList {
		this.conList[i].Close()
	}

	this.conList = make([]Connection, 0)
}

// Id returns the net ID for this ConnectionGroup.
func (this *ConnectionGroup) Id() uint32 {
	return this.id
}

// Key returns the assigned Key for this ConnectionGroup.
func (this *ConnectionGroup) Key() string {
	return this.key
}

// LocalAddr always returns nil for a ConnectionGroup.
func (this *ConnectionGroup) LocalAddr() net.Addr {
	return nil
}

func (this *ConnectionGroup) Name() string {
	return this.name
}

// RemoteAddr always returns nil for a ConnectionGroup.
func (this *ConnectionGroup) RemoteAddr() net.Addr {
	return nil
}

// RemoveConnection removes a connection from the ConnectionGroup.
func (this *ConnectionGroup) RemoveConnection(id uint32) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for i := 0; i < len(this.conList); i++ {
		if this.conList[i].Id() == id {
			this.conList[i] = nil
			this.conList    = append(
				this.conList[:i], 
				this.conList[i + 1:]...,
			)
		}
	}
}

// Send transmits a slice of bytes to member Connections in the
// ConnectionGroup based on the current routing scheme.
func (this *ConnectionGroup) Send(data []byte) {
	if data == nil {
		return
	}
	
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	switch this.routeScheme {
		case BROADCAST:
			for i := range this.conList {
				this.conList[i].Send(data)
			}
		case ROUND_ROBIN:
			this.conList[this.nextConIndex()].Send(data)
	}
}

// nextConIndex increments the currenct connection index, looping through
// all values and resetting back to zero at the end of the list.
func (this *ConnectionGroup) nextConIndex() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	this.lastIndex++

	if this.lastIndex >= len(this.conList) {
		this.lastIndex = 0
	} 

	return this.lastIndex
}
