//  ---------------------------------------------------------------------------
//
//  consolelog.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package log

// Stdlib imports.
import (
    "fmt"
)

// ConsoleLog module name
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
