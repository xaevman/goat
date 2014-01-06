//  ---------------------------------------------------------------------------
//
//  console.go
//
//  Written by Jared Chavez (2013-12-19)
//  Owned by Jared Chavez <xaevman@gmail.com>
//
//  Copyright (c) 2013 Jared Chavez
//
//  -----------

package log

import (
    "fmt"
)

// Module name
const CL_MOD_NAME  = "ConsoleLog"

// Initializes a new ConsoleLog and registers it with the log service.
func InitConsoleLog() {
    consoleLog := new(ConsoleLog)
    RegisterLogSubscriber(consoleLog)
}

// ConsoleLog represents a LogSubscriber responsible for writing logged
// messages to the system console.
type ConsoleLog struct {}

// Crash writes a log message to the system console.
func (this *ConsoleLog) Crash(msg string) {
    fmt.Println(msg)
}

// Debug writes a log message to the sytem console.
func (this *ConsoleLog) Debug(msg string) {
    fmt.Println(msg)
}

// Error writes a log message to the system console.
func (this *ConsoleLog) Error(msg string) {
    fmt.Println(msg)
}

// Flush performs no actions for console services.
func (this *ConsoleLog) Flush() {}

// Info writes a log message to the system console.
func (this *ConsoleLog) Info(msg string) {
    fmt.Println(msg)
}

// Name returns the module name.
func (this *ConsoleLog) Name() string {
    return CL_MOD_NAME
}

// Shutdown performs no actions for console services.
func (this *ConsoleLog) Shutdown() {}
