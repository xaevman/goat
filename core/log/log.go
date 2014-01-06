//  ---------------------------------------------------------------------------
//
//  log.go
//
//  Written by Jared Chavez (2013-12-19)
//  Owned by Jared Chavez <xaevman@gmail.com>
//
//  Copyright (c) 2013 Jared Chavez
//
//  -----------

// Package log implements a standard, multi-channel logging interface.
package log

import (
    "fmt"
    "github.com/xaevman/goat/lib/lifecycle"
    "runtime"
    "sync"
    "time"
    "path/filepath"
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

// Message counters
var (
    crashCount int
    debugCount int
    errorCount int
    infoCount  int
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
        fmt.Println(msg)
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
        fmt.Println(msg)
        return
    }

    debug <- msg
}

// Error formats and logs a message to the error buffer.
func Error(format string, v ...interface{}) {
    msg := getLogMsg(format, "ERROR", v...)

    if !initialized {
        fmt.Println(msg)
        return
    }

    error <- msg
}

// GetLogStats returns a copy of the current log statistics.
func GetLogCounts() LogCounts {
    mutex.Lock()
    defer mutex.Unlock()

    stats := LogCounts{
        Crash: crashCount,
        Debug: debugCount,
        Error: errorCount,
        Info:  infoCount,
    }

    return stats
}

// Info formats and logs a message to the info buffer.
func Info(format string, v ...interface{}) {
    msg := getLogMsg(format, "INFO", v...)

    if !initialized {
        fmt.Println(msg)
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

    crashCount  = 0
    debugCount  = 0
    errorCount  = 0
    infoCount   = 0

    initialized = true
    syncObj     = lifecycle.New()

    mutex.Unlock()

    go async()
}

// RegisterLogSubscriber registers a new log subscriber.
func RegisterLogSubscriber(sub LogSubscriber) {
    mutex.Lock()
    defer mutex.Unlock()

    subscribers[sub.Name()] = sub
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
    Info("LogSubscriber %v unregistered", sub.Name())
}

// async runs in a separate goroutine, forwarding messages in the log buffers
// to registered subscribers. When signaled, async drains all logs and clears
// the list of registered subcribers for a clean shutdown.
func async() {
    for syncObj.QueryRun() {
        select {
        case msg := <- crash:
            sendCrash(msg)
        case msg := <- debug:
            sendDebug(msg)
        case msg := <- error:
            sendError(msg)
        case msg := <- info:
            sendInfo(msg)
        case <-syncObj.QueryShutdown():
        }
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
            time.Now().Format(time.RFC3339),
            log,
            filepath.Base(file),
            line,
            format,
        )
    } else {
        newFmt = fmt.Sprintf(
            "%v [%v] %v",
            time.Now().Format(time.RFC3339),
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

    crashCount++

    for _, v := range subscribers {
        v.Crash(msg)
    }
}

// sendDebug distributes a debug log to all subscribers and increments
// debugCount
func sendDebug(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    debugCount++

    for _, v := range subscribers {
        v.Debug(msg)
    }
}

// sendError distributes an error log to all subscribers and increments
// errorCount
func sendError(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    errorCount++

    for _, v := range subscribers {
        v.Error(msg)
    }
}

// sendInfo distributes an error log to all subscribers and increments
// infoCount
func sendInfo(msg string) {
    mutex.Lock()
    defer mutex.Unlock()

    infoCount++

    for _, v := range subscribers {
        v.Info(msg)
    }
}

// LogCounts encapsulates a snapshot of the number of messages
// logged to each log type so far.
type LogCounts struct {
    Crash int
    Debug int
    Error int
    Info  int
}
