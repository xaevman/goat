//  ---------------------------------------------------------------------------
//
//  disconnect.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// DisconnectChan implements the DisconnectHandler interface and exposes a
// channel based way of receiving disconnected NetIDs from underlying TCPSrv
// and TCPCli instances.
type DisconnectChan struct {
	discoChan chan uint32
}

// NewDisconnectChan is a helper construction function which creates and
// initializes a new DisconnectChan object and returns a pointer to it for use.
func NewDisconnectChan() *DisconnectChan {
	dc := DisconnectChan {
		discoChan: make(chan uint32, 0),
	}

	return &dc
}

// Notify is called by the TCPCli or TCPSrv when a socket disconnects. 
// This should not be used directly by user code. Instead you should 
// insert QueryDisconnect() into your channel based IO logic, where 
// you'll receive signals containing the NetIds of sockets which have 
// been disconnected.
func (this *DisconnectChan) Notify(id uint32) {
	this.discoChan<- id
}

// QueryDisconnect exposes a read-only channel which signals the disconnection
// of an underlying socket. The NetID of the connection is passed over this
// channel.
func (this *DisconnectChan) QueryDisconnect() <-chan uint32 {
	return this.discoChan
}
