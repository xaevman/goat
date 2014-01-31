//  ---------------------------------------------------------------------------
//
//  filelog.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package log

// External imports.
import (
    "github.com/xaevman/goat/lib/fs"
    "github.com/xaevman/goat/lib/lifecycle"
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/lib/time"
)

// Stdlib imports.
import (
    "os"
    "path/filepath"
    stdtime "time"
)

// Perf counters.
const (
    PERF_FLOG_CRASH_BYTES = iota
    PERF_FLOG_DEBUG_BYTES
    PERF_FLOG_ERROR_BYTES
    PERF_FLOG_INFO_BYTES
    PERF_FLOG_FLUSH
    PERF_FLOG_TIMER_FLUSH
    PERF_FLOG_TIMER_IDLE
    PERF_FLOG_COUNT
)

// Perf counter friendly names.
var perfNames = []string {
    "CrashLogSizeBytes",
    "DebugLogSizeBytes",
    "ErrorLogSizeBytes",
    "InfoLogSizeBytes",
    "ManualFlush",
    "TimerFlushhMs",
    "TimerIdleMs",
}

// Default config options
const (
    DEFAULT_BUFFER_DEPTH      = 10000
    DEFAULT_FLUSH_INTERVAL_MS = 1 * 1000
    DEFAULT_LOG_DIR           = "log"
)

// Log file names
const (
    CRASH_LOG_NAME = "crash.log"
    DEBUG_LOG_NAME = "debug.log"
    ERROR_LOG_NAME = "error.log"
    INFO_LOG_NAME  = "info.log"
)

// FileLog module name
const FL_MOD_NAME = "FileLog"


// InitFileLog creates a new FileLog instance, initializes its members,
// registers it with the log service, and spawns a goroutine which is
// responsible for periodically flushing logs to disk.
func InitFileLog() {
    fileLog := FileLog {
        FlushIntervalMs: DEFAULT_FLUSH_INTERVAL_MS,

        crash   : make(chan string, DEFAULT_BUFFER_DEPTH),
        debug   : make(chan string, DEFAULT_BUFFER_DEPTH),
        error   : make(chan string, DEFAULT_BUFFER_DEPTH),
        flush   : make(chan bool,   1),
        info    : make(chan string, DEFAULT_BUFFER_DEPTH),
        perfs   : perf.NewCounterSet(
            "Module.Log." + FL_MOD_NAME,
            PERF_FLOG_COUNT,
            perfNames,
        ),
        syncObj : lifecycle.New(),
    }

    RegisterLogSubscriber(&fileLog)

    Crash("<Log init>")
    Debug("<Log init>")
    Error("<Log init>")
    Info("<Log init>")

    go fileLog.init()
}


// FileLog represents a LogSubscriber which is responsible for
// coordinating writing logged messages to disk.
type FileLog struct {
    FlushIntervalMs int

    crash     chan string
    crashFile *os.File
    debug     chan string
    debugFile *os.File
    error     chan string
    errorFile *os.File
    flush     chan bool
    info      chan string
    infoFile  *os.File
    perfs     *perf.CounterSet
    syncObj   *lifecycle.Lifecycle
}

// Crash writes a log message to the crash log buffer.
func (this *FileLog) Crash(msg string) {
    this.crash <- msg
}

// Debug writes a log message to the debug log buffer.
func (this *FileLog) Debug(msg string) {
    this.debug <- msg
}

// Error writes a log message to the error log buffer.
func (this *FileLog) Error(msg string) {
    this.error <- msg
}

// Flush triggers a log flush, causing messages to be flushed to their
// respective log files.
func (this *FileLog) Flush() {
    this.flush <- true
}

// Info writes a log message to the info log buffer.
func (this *FileLog) Info(msg string) {
    this.info <- msg
}

// Name returns this module's name.
func (this *FileLog) Name() string {
    return FL_MOD_NAME
}

// Shutdown signals the log flush goroutine for shutdown and waits for it
// to finish flushing to disk before returning.
func (this *FileLog) Shutdown() {
    this.syncObj.Shutdown()
}

// flushLogs picks up all buffered log messages and writes them through to 
// their respective files on disk.
func (this *FileLog) flushLogs() {
    for {
        select {
        case msg := <- this.crash:
            this.writeLog(this.crashFile, msg)
        case msg := <- this.debug:
            this.writeLog(this.debugFile, msg)
        case msg := <- this.error:
            this.writeLog(this.errorFile, msg)
        case msg := <- this.info:
            this.writeLog(this.infoFile, msg)
        default:
            return
        }
    }
}

// getFileStats is fired by a timer once per minute and updates the FileLog's
// perf counters with information about log file sizes.
func (this *FileLog) getFileStats() {
    this.perfs.Set(
        PERF_FLOG_CRASH_BYTES, 
        fs.GetFileSize(filepath.Join(DEFAULT_LOG_DIR, CRASH_LOG_NAME)),
    )

    this.perfs.Set(
        PERF_FLOG_DEBUG_BYTES, 
        fs.GetFileSize(filepath.Join(DEFAULT_LOG_DIR, DEBUG_LOG_NAME)),
    )

    this.perfs.Set(
        PERF_FLOG_ERROR_BYTES,
        fs.GetFileSize(filepath.Join(DEFAULT_LOG_DIR, ERROR_LOG_NAME)),
    )

    this.perfs.Set(
        PERF_FLOG_INFO_BYTES,
        fs.GetFileSize(filepath.Join(DEFAULT_LOG_DIR, INFO_LOG_NAME)),
    )

    stdtime.AfterFunc(1 * stdtime.Minute, this.getFileStats)
}

// init runs in a separate goroutine. It ensures that the log directory is
// created, opens the log files for write access, and then responds to timed
// and manual flush requests to write buffered data through to those files. 
// Once signaled for shutdown, init flushes all remaining logs, closes the files
// and signals its completion.
func (this *FileLog) init() {
    stopwatch := new(time.Stopwatch)

    this.perfs.EnableStats(PERF_FLOG_TIMER_FLUSH)
    this.perfs.EnableStats(PERF_FLOG_TIMER_IDLE)

    fs.Mkdir(DEFAULT_LOG_DIR, 0755)

    this.crashFile = this.initLog(CRASH_LOG_NAME)
    this.debugFile = this.initLog(DEBUG_LOG_NAME)
    this.errorFile = this.initLog(ERROR_LOG_NAME)
    this.infoFile  = this.initLog(INFO_LOG_NAME)

    stdtime.AfterFunc(1 * stdtime.Minute, this.getFileStats)

    this.syncObj.StartHeart(this.FlushIntervalMs)

    // run until shutdown
    for this.syncObj.QueryRun() {
        select {
        // manual flush
        case <-this.flush:
            this.perfs.Set(PERF_FLOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            this.flushLogs()
            this.perfs.Set(PERF_FLOG_TIMER_FLUSH, stopwatch.MarkMs())
        // timed flush
        case <-this.syncObj.QueryHeartbeat():
            this.perfs.Set(PERF_FLOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            this.flushLogs()
            this.perfs.Set(PERF_FLOG_TIMER_FLUSH, stopwatch.MarkMs())
        // shutdown
        case <-this.syncObj.QueryShutdown():
            this.perfs.Set(PERF_FLOG_TIMER_IDLE, stopwatch.MarkMs())
        }

        stopwatch.Restart()
    }

    // shutdown
    this.flushLogs()
    this.crashFile.Close()
    this.debugFile.Close()
    this.errorFile.Close()
    this.infoFile.Close()

    this.syncObj.ShutdownComplete()
}

// initLog opens or creates a given log file for append access.
func (this *FileLog) initLog(filePath string) *os.File {
    file, err := fs.AppendFile(filepath.Join(DEFAULT_LOG_DIR, filePath))
    if err != nil {
        Error("Unable to initialize log file %v", filePath)
        this.Shutdown()
        return nil
    }

    return file
}

// writeLog writes the formatted log message msg through to the supplied file
// handle.
func (this *FileLog) writeLog(logFile *os.File, msg string) {
    if logFile == nil {
        return
    }

    logFile.WriteString(msg + "\n")
}
