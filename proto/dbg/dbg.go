//  ---------------------------------------------------------------------------
//
//  dbg.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package dbg implements a protocol which can be included in a goat 
// project to enable realtime performance and diagnostic monitoring over 
// a TCP connection.
package dbg

// External imports.
import (
    "github.com/xaevman/goat/lib/perf"
)


// Perf counters.
const (
    PERF_DBG_ERR_BAD_OBJ = iota
    PERF_DBG_ERR_DESERIALIZE
    PERF_DBG_RCV
    PERF_DBG_SEND
    PERF_DBG_COUNT
)

// Perf counter friendly names
var perfDbgNames = []string {
    "ErrorBadObject",
    "ErrorDeserializeFailed",
    "MessageReceived",
    "MessageSent",
}

// Perf counters.
var dbgPerfs = perf.NewCounterSet(
    "Debug",
    PERF_DBG_COUNT,
    perfDbgNames,
)

// Debugging commands.
const (
    CMD_BLOCKED = iota
    CMD_ENV
    CMD_ERROR
    CMD_MEM
    CMD_PERF
    CMD_RESPONSE
    CMD_STACK
    CMD_SYS
)
