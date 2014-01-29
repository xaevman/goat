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
// into various formats
package diag

import (
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/lib/str"
)

// Stdlib imports.
import (
	"fmt"
	"os"
	"runtime"
)

// Buffer size, in bytes, for storing stack traces.
const TRACE_BUFFER_LEN_B = 1 * 1024 * 1024


// DiagData represents all aggregated diagnostic information.
type DiagData struct {
	System         *SysData
	Environment    *EnvData
	Perfs          *perf.Snapshot
	CallStack      string
	FullStackTrace string
	Memory         *runtime.MemStats
}

// EnvData represents keyval pairs in the environment.
type EnvData struct {
	Vars map[string]string
}

// SysData represents basic information about the system.
type SysData struct {
	Arch     string
	CGOCalls int64
	CPUCount int
	Error    string
	Hostname string
	OS       string
}

// New is a helper function that creates and populates a new DiagData object and
// returns a pointer to it.
func New(err interface{}) *DiagData {
	data := new(DiagData)

	// active stacks
	buffer := make([]byte, TRACE_BUFFER_LEN_B)
	count  := runtime.Stack(buffer, true)
	data.FullStackTrace = string(buffer[:count])

	// call stack
	count = runtime.Stack(buffer, false)
	data.CallStack = string(buffer[:count])

	// environtment
	env     := os.Environ()
	envData := EnvData {
		Vars: make(map[string]string, len(env)),
	}

	for _, v := range env {
		envParts := str.DelimToStrArray(v, "=")
		envData.Vars[envParts[0]] = envParts[1]
	}

	data.Environment = &envData

	// memory
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	data.Memory = stats

	// sys
	hostname, _ := os.Hostname()

	sys := SysData {
		Arch:     runtime.GOARCH,
		CGOCalls: runtime.NumCgoCall(),
		CPUCount: runtime.NumCPU(),
		Error:    fmt.Sprintf("%v", err),
		Hostname: hostname,
		OS:       runtime.GOOS,
	}

	data.System = &sys

	// perf
	data.Perfs = perf.TakeSnapshot()

	return data
}

