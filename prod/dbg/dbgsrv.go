//  ---------------------------------------------------------------------------
//
//  dbgsrv.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package dbg


type Srv struct {
	addr         string
	evtHandler   *net.EventChan
	msgProc      *MsgHandler
	perfs        *perf.CounterSet
	shutdownChan chan bool
	srv          *net.TCPSrv
	syncObj      *lifecycle.Lifecycle
}

type DebugSrv Srv {
	name string
}


