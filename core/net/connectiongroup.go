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

// ConnectionGroup represents a group of Connections, and itself implements 
// the Connection interface so that it can be used interchangebly with single
// Connection objects.
type ConnectionGroup struct {
	conList map[uint32]Connection
	id      uint32
	key     string
	mutex   sync.RWMutex
	name    string
}

// NewConnectionGroup is a constructor helper which builds a newly initalized 
// instance of ConnectionGroup and returns a pointer to it for use.
func NewConnectionGroup(name string) *ConnectionGroup {
	conGroup := ConnectionGroup {
		conList: make(map[uint32]Connection, 0),
		id:      atomic.AddUint32(&netId, 1),
		name:    name,
	}

	return &conGroup
}

// AddConnection adds a new Connection object to this group.
func (this *ConnectionGroup) AddConnection(con Connection) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.conList[con.Id()] = con
}

// Close removes all Connection objects from the group, after calling Close()
// on each.
func (this *ConnectionGroup) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for key, _ := range this.conList {
		this.conList[key].Close()
		delete(this.conList, key)
	}
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

// RemoteAddr always returns nil for a ConnectionGroup.
func (this *ConnectionGroup) RemoteAddr() net.Addr {
	return nil
}

// RemoveConnection removes a connection from the ConnectionGroup.
func (this *ConnectionGroup) RemoveConnection(con Connection) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	delete(this.conList, con.Id())
}

// Send transmits a slice of bytes to all member Connections in the
// ConnectionGroup.
func (this *ConnectionGroup) Send(data []byte) error {
	this.mutex.RLock()
	this.mutex.RUnlock()

	for _, con := range this.conList {
		con.Send(data)
	}

	return nil
}
