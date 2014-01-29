//  ---------------------------------------------------------------------------
//
//  log.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package log implements a standard, multi-channel logging interface.
package log

// External imports.
import (
    "github.com/xaevman/goat/lib/lifecycle"
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/lib/time"
)

// Stdlib imports.
import (
    "fmt"
    "os"
    "runtime"
    "sync"
    stdtime "time"
    "path/filepath"
)

// Perf counters.
const (
    PERF_LOG_BUFFERS= iota
    PERF_LOG_CRASH
    PERF_LOG_DEBUG
    PERF_LOG_ERROR
    PERF_LOG_INFO
    PERF_LOG_INIT
    PERF_LOG_SUBSCRIBER_REG
    PERF_LOG_SUBSCRIBER_UNREG
    PERF_LOG_TIMER_CRASH
    PERF_LOG_TIMER_DEBUG
    PERF_LOG_TIMER_ERROR
    PERF_LOG_TIMER_IDLE
    PERF_LOG_TIMER_INFO
    PERF_LOG_COUNT
)

// Perf counters.
var (
    logPerfNames = []string {
        "Buffers",
        "Crash",
        "Debug",
        "Error",
        "Info",
        "Init",
        "SubscriberRegistered",
        "SubscriberUnregistered",
        "TimerCrashMs",
        "TimerDebugMs",
        "TimerErrorMs",
        "TimerIdleMs",
        "TimerInfoMs",
    }
    logPerfs     = perf.NewCounterSet(
        "Service.Log",
        PERF_LOG_COUNT,
        logPerfNames,
    )
)


// Log write buffers
var (
    crash chan string
    debug chan string
    error chan string
    info  chan string
)

// Enable/Disable debug logging (false by default).
var DebugLogs = false

// Map of log subscribers
var subscribers = map[string]LogSubscriber {}

// Synchronization helpers
var (
    initialized = false
    mutex       sync.Mutex
    syncObj     *lifecycle.Lifecycle
)


// LogSubscriber defines the interface that should be implemented by 
// log subscribers.
type LogSubscriber interface {
    Crash(msg string)
    Debug(msg string)
    Error(msg string)
    Info(msg string)
    Name() string
    Shutdown()
}


// Crash formats and logs a message to the crash buffer.
func Crash(format string, v ...interface{}) {
    msg := getLogMsg(format, "CRASH", v...)

    if !initialized {
        fmt.Fprintln(os.Stderr, msg)
        return
    }

    crash <- msg
}

// Debug formats and logs a message to the debug buffer 
// if debug logging is enabled.
func Debug(format string, v ...interface{}) {
    if !DebugLogs {
        return
    }

    msg := getLogMsg(format, "DEBUG", v...)

    if !initialized {
        fmt.Fprintln(os.Stdout, msg)
        return
    }

    debug <- msg
}

// Error formats and logs a message to the error buffer.
func Error(format string, v ...interface{}) {
    msg := getLogMsg(format, "ERROR", v...)

    if !initialized {
        fmt.Fprintln(os.Stderr, msg)
        return
    }

    error <- msg
}

// Info formats and logs a message to the info buffer.
func Info(format string, v ...interface{}) {
    msg := getLogMsg(format, "INFO", v...)

    if !initialized {
        fmt.Fprintln(os.Stdout, msg)
        return
    }

    info <- msg
}

// Init initializes the logging service, setting up the required channel buffers
// and starting the goroutine which is responsible for sending them to registered 
// subscribers. Subscribers won't receive logs until after Init is called, 
// but message will still be written to the console by default.
func Init(bufferSize int) {
    mutex.Lock()

    crash       = make(chan string, bufferSize)
    debug       = make(chan string, bufferSize)
    error       = make(chan string, bufferSize)
    info        = make(chan string, bufferSize)
    initialized = true
    syncObj     = lifecycle.New()

    mutex.Unlock()

    logPerfs.Set(PERF_LOG_BUFFERS, int64(bufferSize))
    logPerfs.Increment(PERF_LOG_INIT)

    logPerfs.EnableStats(PERF_LOG_TIMER_CRASH)
    logPerfs.EnableStats(PERF_LOG_TIMER_DEBUG)
    logPerfs.EnableStats(PERF_LOG_TIMER_ERROR)
    logPerfs.EnableStats(PERF_LOG_TIMER_IDLE)
    logPerfs.EnableStats(PERF_LOG_TIMER_INFO)

    go async()
}

// RegisterLogSubscriber registers a new log subscriber.
func RegisterLogSubscriber(sub LogSubscriber) {
    mutex.Lock()
    defer mutex.Unlock()

    subscribers[sub.Name()] = sub

    logPerfs.Increment(PERF_LOG_SUBSCRIBER_REG)
    Info("LogSubscriber %v registered", sub.Name())
}

// Shutdown disables the run flag and waits for log buffers to empty
// before returning.
func Shutdown() {
    if !initialized {
        return
    }

    initialized = false
    
    syncObj.Shutdown()
}

// UnregisterLogSubscriber removes a log subscriber.
func UnregisterLogSubscriber(sub LogSubscriber) {
    mutex.Lock()
    defer mutex.Unlock()

    delete(subscribers, sub.Name())
    logPerfs.Increment(PERF_LOG_SUBSCRIBER_UNREG)
    Info("LogSubscriber %v unregistered", sub.Name())
}

// async runs in a separate goroutine, forwarding messages in the log buffers
// to registered subscribers. When signaled, async drains all logs and clears
// the list of registered subcribers for a clean shutdown.
func async() {
    stopwatch := new(time.Stopwatch)

    for syncObj.QueryRun() {
        select {
        case msg := <- crash:
            logPerfs.Set(PERF_LOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            sendCrash(msg)
            logPerfs.Set(PERF_LOG_TIMER_CRASH, stopwatch.MarkMs())
        case msg := <- debug:
            logPerfs.Set(PERF_LOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            sendDebug(msg)
            logPerfs.Set(PERF_LOG_TIMER_DEBUG, stopwatch.MarkMs())
        case msg := <- error:
            logPerfs.Set(PERF_LOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            sendError(msg)
            logPerfs.Set(PERF_LOG_TIMER_ERROR, stopwatch.MarkMs())
        case msg := <- info:
            logPerfs.Set(PERF_LOG_TIMER_IDLE, stopwatch.MarkMs())

            stopwatch.Restart()
            sendInfo(msg)
            logPerfs.Set(PERF_LOG_TIMER_INFO, stopwatch.MarkMs())
        case <-syncObj.QueryShutdown():
            logPerfs.Set(PERF_LOG_TIMER_IDLE, stopwatch.MarkMs())
        }

        stopwatch.Restart()
    }

    Info("Shutdown initiated")
    drainLogs()
    clearRegistrations()

    syncObj.ShutdownComplete()
}

// clearRegistrations clears the list of registered subsribers, calling Shutdown()
// on each of them along the way.
func clearRegistrations() {
    mutex.Lock()
    defer mutex.Unlock()

    for _, v := range subscribers {
        v.Shutdown()
    }

    for _, v := range subscribers {
        delete(subscribers, v.Name())
    }
}

// drainLogs is called during shutdown to drain all channel buffers and return.
func drainLogs() {
    for {
        select {
        case msg := <- crash:
            sendCrash(msg)
        case msg := <- debug:
            sendDebug(msg)
        case msg := <- error:
            sendError(msg)
        case msg := <- info:
            sendInfo(msg)
        default:
            return
        }
    }
}

// getLogMsg formats a log message with the specified format string and variadic
// arguments, adding a time stamp, the log type, and file line (if possible) along
// the way.
func getLogMsg(format string, log string, v ...interface{}) string {
    var newFmt string

    _, file, line, ok := runtime.Caller(2)

    if ok {
        newFmt = fmt.Sprintf(
            "%v [%v] <%v:%v> %v",
            stdtime.Now().Format(stdtime.RFC3339),
            log,
            filepath.Base(file),
            line,
            format,
        )
    } else {
        newFmt = fmt.Sprintf(
            "%v [%v] %v",
            stdtime.Now().Format(stdtime.RFC3339),
            log,
            format,
        )
    }

    return fmt.Sprintf(newFmt, v...)
}

// sendCrash distributes a crash log to all subscribers and increments
// crashCount
func sendCrash(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    for _, v := range subscribers {
        v.Crash(msg)
    }

    logPerfs.Increment(PERF_LOG_CRASH)
}

// sendDebug distributes a debug log to all subscribers and increments
// debugCount
func sendDebug(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    for _, v := range subscribers {
        v.Debug(msg)
    }

    logPerfs.Increment(PERF_LOG_DEBUG)
}

// sendError distributes an error log to all subscribers and increments
// errorCount
func sendError(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    for _, v := range subscribers {
        v.Error(msg)
    }

    logPerfs.Increment(PERF_LOG_ERROR)
}

// sendInfo distributes an error log to all subscribers and increments
// infoCount
func sendInfo(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    for _, v := range subscribers {
        v.Info(msg)
    }

    logPerfs.Increment(PERF_LOG_INFO)
}
