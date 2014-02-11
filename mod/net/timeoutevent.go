//  ---------------------------------------------------------------------------
//
//  timeoutevent.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// Timeout types.
const (
    TIMEOUT_CONNECT = iota
    TIMEOUT_DISCONNECT
    TIMEOUT_GENERAL
    TIMEOUT_RCV
    TIMEOUT_SEND
)

// TimeoutEvent represents a network operation and associated data which has
// timed out. TimeoutEvents can be handled or ignored by user code as is
// appropriate for the given application.
type TimeoutEvent struct {
    Data        interface{}
    MessageType uint16
    ParentId    uint32
    TimeoutType int
}
