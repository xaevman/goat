//  ---------------------------------------------------------------------------
//
//  connectiongroup.go
//
//  Copyright (c) 2013, Jared Chavez. 
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
)


type ConnectionGroup struct {
	conList map[uint32]Connection
	id      uint32
	key     string
	mutex   sync.RWMutex
}

func (this *ConnectionGroup) AddConnection(con Connection) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.conList[con.Id()] = con
}

func (this *ConnectionGroup) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for key, _ := range this.conList {
		this.conList[key].Close()
		delete(this.conList, key)
	}
}

func (this *ConnectionGroup) Id() uint32 {
	return this.id
}

func (this *ConnectionGroup) Key() string {
	return this.key
}

func (this *ConnectionGroup) LocalAddr() net.Addr {
	return nil
}

func (this *ConnectionGroup) RemoteAddr() net.Addr {
	return nil
}

func (this *ConnectionGroup) RemoveConnection(con Connection) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	delete(this.conList, con.Id())
}

func (this *ConnectionGroup) Send(data []byte) error {
	this.mutex.RLock()
	this.mutex.RUnlock()

	for _, con := range this.conList {
		con.Send(data)
	}

	return nil
}
