//  ---------------------------------------------------------------------------
//
//  diagstr.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package diag

import (
	"bytes"
	"fmt"
)

// AsString aggregates and returns text-based diagnostics information.
// Diagnostic information includes hostname, CPU count, environment data,
// stack traces for all running goroutines, and memory allocation statistics.
func AsString(err interface{}) string {
	data := New(err)

	return fmt.Sprintf(
		"\n==== [Begin Crash Report] ====\n"+
			"Error:    %v\n"+
			"Hostname: %v\n"+
			"CPUCount: %v\n"+
			"CGOCalls: %v\n"+
			"GOOS:     %v\n"+
			"GOARCH:   %v\n\n"+
			"%v"+
			"%v"+
			"%v"+
			"==== [End Crash Report] ====",
		data.System.Error,
		data.System.Hostname,
		data.System.CPUCount,
		data.System.CGOCalls,
		data.System.OS,
		data.System.Arch,
		fmtEnvStr(data),
		fmtStackTraceStr(data),
		fmtMemStatsStr(data),
	)
}

// fmtEnvStr dumps the current environment data and formats it as text.
func fmtEnvStr(data *DiagData) string {
	var buffer bytes.Buffer

	buffer.WriteString("==== [Begin Env] ====\n")

	for k, v := range data.Environment.Vars {
		buffer.WriteString(fmt.Sprintf("%v = %v\n", k, v))
	}

	buffer.WriteString("==== [End Env] ====\n\n")

	return buffer.String()
}

// fmtMemStatsStr dumps memory allocation statistics and formats it as text.
func fmtMemStatsStr(data *DiagData) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf(
		"==== [Begin MemStats] ====\n"+
		"Alloc:        %v\n"+
		"TotalAlloc:   %v\n"+
		"Sys:          %v\n"+
		"Lookups:      %v\n"+
		"Mallocs:      %v\n"+
		"Frees:        %v\n"+
		"HeapAlloc:    %v\n"+
		"HeapSys:      %v\n"+
		"HeapIdle:     %v\n"+
		"HeapInuse:    %v\n"+
		"HeapReleased: %v\n"+
		"HeapObjects:  %v\n"+
		"StackInuse:   %v\n"+
		"StackSys:     %v\n"+
		"MSpanInuse:   %v\n"+
		"MSpanSys:     %v\n"+
		"MCacheInuse:  %v\n"+
		"MCacheSys:    %v\n"+
		"BuckHashSys:  %v\n"+
		"GCSys:        %v\n"+
		"OtherSys:     %v\n"+
		"NextGC:       %v\n"+
		"LastGC:       %v\n"+
		"PauseTotalNs: %v\n"+
		"NumGC:        %v\n"+
		"EnableGC:     %v\n"+
		"DebugGC:      %v\n"+
		"==== [End MemStats] ====\n\n",
		data.Memory.Alloc,
		data.Memory.TotalAlloc,
		data.Memory.Sys,
		data.Memory.Lookups,
		data.Memory.Mallocs,
		data.Memory.Frees,
		data.Memory.HeapAlloc,
		data.Memory.HeapSys,
		data.Memory.HeapIdle,
		data.Memory.HeapInuse,
		data.Memory.HeapReleased,
		data.Memory.HeapObjects,
		data.Memory.StackInuse,
		data.Memory.StackSys,
		data.Memory.MSpanInuse,
		data.Memory.MSpanSys,
		data.Memory.MCacheInuse,
		data.Memory.MCacheSys,
		data.Memory.BuckHashSys,
		data.Memory.GCSys,
		data.Memory.OtherSys,
		data.Memory.NextGC,
		data.Memory.LastGC,
		data.Memory.PauseTotalNs,
		data.Memory.NumGC,
		data.Memory.EnableGC,
		data.Memory.DebugGC,
	))

	buffer.WriteString("==== [Begin Malloc Stats] ====\n")

	for i, v := range data.Memory.BySize {
		if v.Mallocs < 1 {
			continue
		}

		memSize := fmt.Sprintf(
			"\tSize:    %v\n"+
			"\tMallocs: %v\n"+
			"\tFrees:   %v\n",
			v.Size,
			v.Mallocs,
			v.Frees,
		)

		buffer.WriteString(memSize)

		if i < len(data.Memory.BySize)-1 {
			buffer.WriteString("\n")
		}
	}

	buffer.WriteString("==== [End Malloc Stats] ====\n\n")

	return buffer.String()
}

// fmtStackTraceStr dumps the current stack, and the stack of any executing
// goroutines and formats the data as text.
func fmtStackTraceStr(data *DiagData) string {
	stackTrace := fmt.Sprintf(
		"==== [Begin Stack Trace] ====\n"+
		"%v"+
		"==== [End Stack Trace] ====\n\n",
		data.CallStack,
	)

	fullTrace := fmt.Sprintf(
		"==== [Begin Full Stack Trace] ====\n"+
		"%v"+
		"==== [End Full Stack Trace] ====\n\n",
		data.FullStackTrace,
	)

	return stackTrace + fullTrace
}
