//  ---------------------------------------------------------------------------
//
//  loop.go
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
    "github.com/xaevman/goat/mod/log"
)

// Stdlib imports.
import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
)

// GoGraphLoop is a goapp.LoopHanlder implementation for a GoGraph
// instance.
type GoGraphLoop struct {}

// OnHeartbeat captures the timestamp of this heartbeat, checks all
// configured stats, and attempts to upload their current values to
// the configured Graphite server.
func (this *GoGraphLoop) OnHeartbeat() {
    timestamp := time.Now()

    for i := range sysCtlStats {
        sendStat(sysCtlStats[i], timestamp)
    }
}

// PreLoop is unused in GoGraph.
func (this *GoGraphLoop) PreLoop() {}

// PostLoop is unused in GoGraph.
func (this *GoGraphLoop) PostLoop() {}

// sendStat queries for a given stat, formats it in the correct format,
// and attempts to send it over the open socket to Graphite.
func sendStat (stat string, timestamp time.Time) {
    val, err := getStat(stat)
    if err != nil {
        log.Error(err.Error())
        return
    }

    host, err := os.Hostname()
    if err != nil {
        log.Error(err.Error())
        host = "unknown"
    }

    host = strings.Replace(host, ".", "_", -1)

    graphTxt := fmt.Sprintf(
        GRAPH_MSG_FORMAT,
        statPrefix,
        host,
        stat,
        val,
        timestamp.Unix(),
    )

    _, err = srvCon.Write([]byte(graphTxt))
    if err != nil {
        log.Error(err.Error())
        return
    }

    log.Debug(strings.TrimSpace(graphTxt))
}

// getStat attempts to query sysctl for the provided stat
// and return its value.
func getStat (stat string) (string, error) {
    cmd    := exec.Command("sysctl", stat)
    out, _ := cmd.Output()

    outTxt   := strings.TrimSpace(string(out))
    outParts := strings.SplitN(outTxt, ":", 2)

    if len(outParts) < 2 {
        errTxt := fmt.Sprintf("stat parse failure: %s", stat)
        return "", errors.New(errTxt)
    }

    val := strings.TrimSpace(outParts[1])

    return val, nil
}
