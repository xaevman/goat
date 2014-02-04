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
	"bytes"
	"fmt"
	"os"
	"runtime"
)

// Buffer size, in bytes, for storing stack traces.
const TRACE_BUFFER_LEN_B = 1 * 1024 * 1024 // 1MB


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

// String pretty-prints the EnvData object.
func (this *EnvData) String() string {
	var buffer bytes.Buffer

	for k, v := range this.Vars {
		buffer.WriteString(fmt.Sprintf("%v = %v\n", k, v))
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
	data.FullStackTrace = NewFullStackTrace()

	// call stack
	data.CallStack = NewStackTrace()

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

// NewStackTrace dumps the current calling stack in string format.
func NewStackTrace() string {
	buffer := make([]byte, TRACE_BUFFER_LEN_B)
	count  := runtime.Stack(buffer, false)	
	return string(buffer[:count])
}

// NewFullStackTrace dumps stack traces for all active go routines.
func NewFullStackTrace() string {
	buffer := make([]byte, TRACE_BUFFER_LEN_B)
	count  := runtime.Stack(buffer, true)
	return string(buffer[:count])
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

// NewMemData creates and populates a new runtime.MemStats object and
// returns a pointer to it for use.
func NewMemData() *runtime.MemStats {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return stats
}

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
