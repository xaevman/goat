//  ---------------------------------------------------------------------------
//
//  broadcastgroup.go
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

// BroadcastGroup represents a group of Connections that can be addressed
// by via a single NetID. BroadcastGroup itself implements the Connection 
// interface so that it can be used interchangebly with single Connection 
// objects.
type BroadcastGroup struct {
	conList     map[uint32]Connection
	id          uint32
	key         string
	lastIndex   int
	mutex       sync.RWMutex
	name        string
}

// NewBroadcastGroup is a constructor helper which builds a newly initalized 
// instance of BroadcastGroup and returns a pointer to it for use.
func NewBroadcastGroup(name string) *BroadcastGroup {
	conGroup := BroadcastGroup {
		conList: make(map[uint32]Connection, 0),
		id:      atomic.AddUint32(&netId, 1),
		name:    name,
	}

	return &conGroup
}

// AddConnection adds a new Connection object to this group.
func (this *BroadcastGroup) AddConnection(con Connection) {
	if con == nil {
		return
	}

	this.mutex.RLock()
	if this.conList[con.Id()] != nil {
		this.mutex.RUnlock()
		return
	}
	this.mutex.RUnlock()

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.conList[con.Id()] = con
}

// Close removes all Connection objects from the group, after calling Close()
// on each.
func (this *BroadcastGroup) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for k, _ := range this.conList {
		this.conList[k].Close()
	}

	this.conList = make(map[uint32]Connection, 0)
}

// GetConnection returns the member Connection object matching the given id
// if one exists, otherwise returns nil.
func (this *BroadcastGroup) GetConnection(id uint32) Connection {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.conList[id]
}

// Id returns the net ID for this BroadcastGroup.
func (this *BroadcastGroup) Id() uint32 {
	return this.id
}

// Key returns the assigned Key for this BroadcastGroup.
func (this *BroadcastGroup) Key() string {
	return this.key
}

// LocalAddr always returns nil for a BroadcastGroup.
func (this *BroadcastGroup) LocalAddr() net.Addr {
	return nil
}

func (this *BroadcastGroup) Name() string {
	return this.name
}

// RemoteAddr always returns nil for a BroadcastGroup.
func (this *BroadcastGroup) RemoteAddr() net.Addr {
	return nil
}

// RemoveConnection removes a connection from the BroadcastGroup.
func (this *BroadcastGroup) RemoveConnection(id uint32) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	delete(this.conList, id)
}

// Send transmits a slice of bytes to member Connections in the
// BroadcastGroup.
func (this *BroadcastGroup) Send(data []byte) {
	if data == nil {
		return
	}
	
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	for k, _ := range this.conList {
		this.conList[k].Send(data)
	}
}
