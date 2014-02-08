//  ---------------------------------------------------------------------------
//
//  diagstr.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package diag

// Stdlib imports.
import (
    "bytes"
    "fmt"
    "runtime"
)

// AsString aggregates and returns text-based diagnostics information.
// Diagnostic information includes hostname, CPU count, environment data,
// stack traces for all running goroutines, and memory allocation statistics.
func AsString(diagData *DiagData) string {
    data := New()

    return fmt.Sprintf(
        "\n"                                   +
        "==== [Begin Crash Report] ====\n"     +
        "%s\n\n"                               +
        "==== [Begin Env] ====\n"              + 
        "%s"                                   +
        "==== [End Env] ====\n\n"              +
        "==== [Begin Perf Snapshot] ====\n"    +
        "%s"                                   +
        "==== [End Perf Snapshot] ====\n\n"    +
        "==== [Begin Stack Trace] ====\n"      +
        "%s"                                   +
        "==== [End Stack Trace] ====\n\n"      +
        "==== [Begin Full Stack Trace] ====\n" +
        "%s"                                   +
        "==== [End Full Stack Trace] ====\n\n" +
        "==== [Begin MemStats] ====\n"         +
        "%s"                                   +
        "==== [End MemStats] ====\n\n"         +
        "==== [Begin Malloc Stats] ====\n"     +
        "%s"                                   +
        "==== [End Malloc Stats] ====\n\n"     +
        "==== [End Crash Report] ====",
        data.System,
        data.Environment,
        data.Perfs,
        data.CallStack,
        data.FullStackTrace,
        FmtMemStatsStr(data.Memory),
        FmtMallocStatsStr(data.Memory),
    )
}


// FmtMemStatsStr dumps memory allocation statistics and formats it as text.
// Malloc information is not included for brevity. Get it by calling
// FmtMallocStatsStr.
func FmtMemStatsStr(data *runtime.MemStats) string {
    return fmt.Sprintf(
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
        "DebugGC:      %v\n",
        data.Alloc,
        data.TotalAlloc,
        data.Sys,
        data.Lookups,
        data.Mallocs,
        data.Frees,
        data.HeapAlloc,
        data.HeapSys,
        data.HeapIdle,
        data.HeapInuse,
        data.HeapReleased,
        data.HeapObjects,
        data.StackInuse,
        data.StackSys,
        data.MSpanInuse,
        data.MSpanSys,
        data.MCacheInuse,
        data.MCacheSys,
        data.BuckHashSys,
        data.GCSys,
        data.OtherSys,
        data.NextGC,
        data.LastGC,
        data.PauseTotalNs,
        data.NumGC,
        data.EnableGC,
        data.DebugGC,
    )
}

// FmtMallocStatsStr outputs the extended malloc information from a given
// MemStats instance as basic text.
func FmtMallocStatsStr(data *runtime.MemStats) string {
    var buffer bytes.Buffer

    for i, v := range data.BySize {
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

        if i < len(data.BySize)-1 {
            buffer.WriteString("\n")
        }
    }

    return buffer.String()
}
