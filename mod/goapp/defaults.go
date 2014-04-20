//  ---------------------------------------------------------------------------
//
//  defaults.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package goapp

// External imports.
import (
    "github.com/xaevman/goat/mod/diag"
    "github.com/xaevman/goat/mod/log"
)


// DefaultAppStarter is the default AppStarter implementation for a GoApp
// unless overwridden via SetAppStarter().
type DefaultAppStarter struct {}

// PreInit logs a debug message to the log service.
func (this *DefaultAppStarter) PreInit() {
    log.Debug("PreInit")
}

// PostInit logs a debug message to the log service.
func (this *DefaultAppStarter) PostInit() {
    log.Debug("PostInit")
}


// DefaultAppCloser is the default AppCloser implementation for a GoApp
// unless overwridden via SetAppCloser().
type DefaultAppCloser struct {}

// PreShutdown logs a debug message to the log service.
func (this *DefaultAppCloser) PreShutdown() {
    log.Debug("PreShutdown")
}

// PostShutdown logs a debug message to the log service.
func (this *DefaultAppCloser) PostShutdown() {
    log.Debug("PostShutdown")
}


// DefaultCrashHandler is the default CrashHandler implementation for a GoApp
// unless overwridden via SetCrashHandler().
type DefaultCrashHandler struct {}

// OnCrash logs a crash message to the log service and then calls panic with
// the same panic data.
func (this *DefaultCrashHandler) OnCrash(crashData interface{}) {
    diagData := diag.New()

    log.Error("%s", crashData)
    log.Crash(diag.AsString(diagData))

    panic(crashData)
}


// DefaultLoopHandler is teh default LoopHandler implementation for a GoApp
// unless overwridden via SetLoopHandler().
type DefaultLoopHandler struct {}

// OnHeartbeat logs a debug message to the log service.
func (this *DefaultLoopHandler) OnHeartbeat() {
    log.Debug("OnHeartbeat")
}

// PreLoop logs a debug message to the log service.
func (this *DefaultLoopHandler) PreLoop() {
    log.Debug("PreLoop")
}

// PostLoop logs a debug message to the log service.
func (this *DefaultLoopHandler) PostLoop() {
    log.Debug("PostLoop")
}
