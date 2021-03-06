//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package log

import(
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "testing"
)

// Test config options.
var (
    testPath = filepath.Join(".", "log")
    runs     = 10000
)

// TestConsoleLog initializes the ConsoleLog and writes a few thousand
// logs. The test checks to make sure that init, log writing and shutdown
// proceed without error.
func TestLogsWithDebug(t *testing.T) {
    // init
    DebugLogs = true

    logPerfs.Reset()

    cCount, dCount, eCount, iCount := doLogRun()

    // Actual counts should be at least as high as the number sent into
    // the system. They will often be higher due to automatic system
    // log entries
    if  logPerfs.Value(PERF_LOG_CRASH) < cCount || 
        logPerfs.Value(PERF_LOG_DEBUG) < dCount ||
        logPerfs.Value(PERF_LOG_ERROR) < eCount || 
        logPerfs.Value(PERF_LOG_INFO)  < iCount {
        t.Fatalf(
            "TestLogsWithDebug failed, count mismatch\n" +
            "\tcrash (%v vs %v)\n" +
            "\tdebug (%v vs %v)\n" +
            "\terror (%v vs %v)\n" +
            "\tinfo  (%v vs %v)\n",
            logPerfs.Value(PERF_LOG_CRASH), cCount,
            logPerfs.Value(PERF_LOG_DEBUG), dCount,
            logPerfs.Value(PERF_LOG_ERROR), eCount,
            logPerfs.Value(PERF_LOG_INFO),  iCount,
        )
    }

    // validate file lines
    validateFileLines(
        filepath.Join(DEFAULT_LOG_DIR, CRASH_LOG_NAME), 
        logPerfs.Value(PERF_LOG_CRASH),
        t,
    )
    validateFileLines(
        filepath.Join(DEFAULT_LOG_DIR, DEBUG_LOG_NAME), 
        logPerfs.Value(PERF_LOG_DEBUG),
        t,
    )
    validateFileLines(
        filepath.Join(DEFAULT_LOG_DIR, ERROR_LOG_NAME), 
        logPerfs.Value(PERF_LOG_ERROR),
        t,
    )
    validateFileLines(
        filepath.Join(DEFAULT_LOG_DIR, INFO_LOG_NAME), 
        logPerfs.Value(PERF_LOG_INFO),
        t,
    )

    fmt.Println("TestLogsWithDebug: passed")
}

// TestLogsNoDebug does a test run without debug logging enabled,
// then checking the log counts to make sure no debug logs were dispatched.
func _TestLogsNoDebug(t *testing.T) {
    // init
    DebugLogs = false

    logPerfs.Reset()

    doLogRun()

    if logPerfs.Value(PERF_LOG_DEBUG) != 0 {
        t.Fatalf(
            "TestLogsNoDebug failed: debug count %d",
            logPerfs.Value(PERF_LOG_DEBUG) ,
        )
    }

    fmt.Println("TestLogsNoDebug: passed")
}

// doLogRun performs a log run with the FileLog and ConsoleLog providers and
// returns the number of each type of message that was passed into the dispatcher.
func doLogRun() (cCount, dCount, eCount, iCount int64) {
    os.RemoveAll(testPath)
    
    Init(100)

    // Important to initialize FileLog first, or testing file line counts
    // will fail due to messages making it through the buffer to ConsoleLog
    // before FileLog is able to fully initialize.
    InitFileLog()
    InitConsoleLog()

    // do some work
    r := rand.New(rand.NewSource(10))
    for i := 0; i < runs; i++ {
        num := r.Intn(4)
        switch num {
        case 0:
            cCount++
            Crash("counter: %v", i)
        case 1:
            dCount++
            Debug("counter: %v", i)
        case 2:
            eCount++
            Error("counter: %v", i)
        case 3:
            iCount++
            Info("counter: %v", i)
        }
    }

    // shutdown
    Shutdown()

    return
}

// validateFileLines checks to make sure that appropriate file logs were
// written, and that they have the same number of messages as were sent to 
// them by the log dispatcher.
func validateFileLines(path string, count int64, t *testing.T) {
    file, err := os.OpenFile(path, os.O_RDONLY, 0750)
    if err != nil {
        t.Fatal(err)
    }
    defer file.Close()

    var fCount int64 = 0
    reader := bufio.NewReader(file)

    for {
        _, err := reader.ReadString('\n')
        if err != nil {
            break   // eof
        }

        fCount++
    }

    if fCount != count {
        t.Fatalf(
            "validateFileLines failed on file %v\n" +
            "\tfCount(%v) != count(%v)",
            path,
            fCount,
            count,
        )
    }
}

// TestCleanup cleans up temporary log directory after a test run.
func TestCleanup(t *testing.T) {
    err := os.RemoveAll(testPath)
    if err != nil {
        t.Fatal(err)
    }
}
