//  ---------------------------------------------------------------------------
//
//  start.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package main

// External imports.
import (
    "github.com/xaevman/goat/mod/goapp"
    "github.com/xaevman/goat/mod/config"
    "github.com/xaevman/goat/mod/log"
)

// Stdlib imports.
import (
    "net"
    "strings"
)

// GoGraphStart is a goapp.AppStarter implementation for a GoGraph
// instance.
type GoGraphStart struct {}

// PreInit registers the application's ini config and queries to
// determine if debug logs should be enabled or not.
func (this *GoGraphStart) PreInit() {
    config.InitIniProvider("config/gograph.ini", 1)
    debugLogs, _ := config.GetBoolVal("System.DebugLogs", 0, false)
    log.DebugLogs = debugLogs
}

// PostInit queries the registered config to determine what prefix
// to append to stat names, the address and port of the Graphite 
// server, the interval at which to collect statistics, and which
// statistics to look for.
func (this *GoGraphStart) PostInit() {
    // Stat prefix
    prefix, _ := config.GetVal(
        "Graphite.StatPrefix",
        0,
        "srv",
    )
    statPrefix = prefix

    // Server address
    srvAddr, _ := config.GetVal(
        "Graphite.SrvAddr", 
        0, 
        "127.0.0.1:2003",
    )
    
    // Send interval (in ms)
    heartbeatMs, _ := config.GetIntVal(
        "Graphite.SendIntervalMs",
        0,
        60000,
    )

    // Sysctl stats to gather
    sStatEntries := config.GetEntries("Stats.SysctlStat")

    for _, entry := range sStatEntries {
        cleanVal := strings.TrimSpace(entry.GetVal(0))
        if cleanVal == "" {
            continue
        }

        sysCtlStats = append(sysCtlStats, cleanVal)
    }

    // Apply config
    con, err := net.Dial("tcp", srvAddr)
    if err != nil {
        log.Error("%v", err)
        goapp.SetExitCode(1)
        goapp.Stop()
        return
    }
    srvCon = con

    goapp.SetHeartbeat(heartbeatMs)

    log.Info(
        "Initialized (%s, heartbeat %dms)", 
        srvAddr, 
        heartbeatMs,
    )
}

