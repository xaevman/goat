//  ---------------------------------------------------------------------------
//
//  diag.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package diag implements functions for dumping crash information
// into various formats.
package diag

import (
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/lib/str"
)

// Stdlib imports.
import (
    "bytes"
    "fmt"
    "os"
    "regexp"
    "runtime"
    "sort"
    "strings"
)


// Buffer size, in bytes, for storing stack traces.
const TRACE_BUFFER_LEN_B = 1 * 1024 * 1024 // 1MB

// regex for parsing out the start of blocked go routines
// in stack dumps.
var blockedRegex = regexp.MustCompile(
    "(?m)goroutine \\d* \\[chan .*\\]\\:$",
)


// DiagData represents all aggregated diagnostic information.
type DiagData struct {
    System         *SysData
    Environment    *EnvData
    Perfs          *perf.Snapshot
    StackTrace     []*StackTrace
    Memory         *runtime.MemStats
}


// EnvData represents keyval pairs in the environment.
type EnvData struct {
    Vars map[string]string
}

// String pretty-prints the EnvData object.
func (this *EnvData) String() string {
    var buffer bytes.Buffer

    count := 0
    sKeys := make([]string, len(this.Vars))
    for k, _ := range this.Vars {
        sKeys[count] = k
        count++
    }

    sort.Strings(sKeys)

    for i := range sKeys {
        buffer.WriteString(fmt.Sprintf(
            "%v = %v\n", 
            sKeys[i], 
            this.Vars[sKeys[i]],
        ))
    }

    return buffer.String()
}


// SysData represents basic information about the system.
type SysData struct {
    Arch        string
    CGOCalls    int64
    CPUCount    int
    Error       string
    GoRoutines  int
    GoVersion   string
    Hostname    string
    OS          string
}

// String pretty-prints the SysData object.
func (this *SysData) String() string {
    return fmt.Sprintf(
        "Error:        %v\n"   +
        "Hostname:     %v\n"   +
        "CPUCount:     %v\n"   +
        "CGOCalls:     %v\n"   +
        "OS:           %v\n"   +
        "Architecture: %v\n"   +
        "Go Routines:  %v\n"   +
        "Go Version:   %v\n",
        this.Error,
        this.Hostname,
        this.CPUCount,
        this.CGOCalls,
        this.OS,
        this.Arch,
        this.GoRoutines,
        this.GoVersion,
    )
}

// New is a helper function that creates and populates a new DiagData object and
// returns a pointer to it.
func New() *DiagData {
    data := new(DiagData)

    // active stacks
    data.StackTrace = NewStackTrace()

    // environtment
    data.Environment = NewEnvData()

    // memory
    data.Memory = NewMemData()

    // sys
    data.System = NewSysData()

    // perf
    data.Perfs = perf.TakeSnapshot()

    return data
}

// NewBlockData returns call stack information for any go routines
// which are currently blocking.
func NewBlockedData() string {
    var reading bool
    var results bytes.Buffer
    var temp    bytes.Buffer

    readBuffer := bytes.NewBufferString(NewStackString())
    line, err  := readBuffer.ReadString('\n');

    for err == nil {
        if reading && strings.TrimSpace(line) == "" {
            reading = false
            temp.WriteString(line)
            results.WriteString(temp.String())
            temp.Reset()

            line, err = readBuffer.ReadString('\n')

            continue
        }

        if reading {
            temp.WriteString(line)

            line, err = readBuffer.ReadString('\n')

            continue
        }

        if blockedRegex.MatchString(line) {
            temp.WriteString(line)
            reading = true
        }

        line, err = readBuffer.ReadString('\n')
    }

    return results.String()
}

// NewEnvData creates and populates a new EnvData object and returns
// a pointer to it for use.
func NewEnvData() *EnvData {
    env     := os.Environ()
    envData := EnvData {
        Vars: make(map[string]string, len(env)),
    }

    for _, v := range env {
        envParts := str.DelimToStrArray(v, "=")
        envData.Vars[envParts[0]] = envParts[1]
    }

    return &envData
}

// NewStackString dumps stack traces for all active go routines.
func NewStackString() string {
    buffer := make([]byte, TRACE_BUFFER_LEN_B)
    count  := runtime.Stack(buffer, true)
    return string(buffer[:count])
}

// NewMemData creates and populates a new runtime.MemStats object and
// returns a pointer to it for use.
func NewMemData() *runtime.MemStats {
    stats := new(runtime.MemStats)
    runtime.ReadMemStats(stats)
    return stats
}

// // NewStackTrace dumps the current calling stack in string format.
// func NewStackTrace() string {
//     buffer := make([]byte, TRACE_BUFFER_LEN_B)
//     count  := runtime.Stack(buffer, false)  
//     return string(buffer[:count])
// }

// NewSysData creatse and populates a new SysData object and returns a
// pointer to it for use.
func NewSysData() *SysData {
    hostname, _ := os.Hostname()

    sys := SysData {
        Arch        : runtime.GOARCH,
        CGOCalls    : runtime.NumCgoCall(),
        CPUCount    : runtime.NumCPU(),
        Hostname    : hostname,
        GoRoutines  : runtime.NumGoroutine(),
        GoVersion   : runtime.Version(),
        OS          : runtime.GOOS,
    }

    return &sys
}
