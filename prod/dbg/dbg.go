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

// Package dbg implements a protocol, server and client structure that can
// be included in a goat project to enable realtime performance and diagnostic
// monitoring over a TCP connection.
package dbg

// External imports.
import (
	"github.com/xaevman/goat/core/net"
)

// Network protocol object.
var Protocol *net.Protocol

