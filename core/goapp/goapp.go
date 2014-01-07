//  ---------------------------------------------------------------------------
//
//  goapp.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package goapp
package goapp

// External imports.
import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/lifecycle"
)

// Stdlib imports.
import (
	"sync"
)

// Application properties.
var	(
	appName       string
	initialized = false
)

// Synchronization helpers.
var (
	msgPump = make(chan bool, 1)
	mutex     sync.Mutex
	syncObj = lifecycle.New()
)

// Application interface instances.
var (
	appStarter   AppStarter   = new(DefaultAppStarter)
	appCloser    AppCloser    = new(DefaultAppCloser)
	crashHandler CrashHandler = new(DefaultCrashHandler)
	loopHandler  LoopHandler  = new(DefaultLoopHandler)
)

// AppStarter defines the interface which should be implemented
// and registered via SetAppStarter() to execute user code before
// and just after application initialization. 
type AppStarter interface {
	PreInit()
	PostInit()
}

// AppCloser defines the interface which should be implemented
// and registered via SetAppCloser() to execute user code just before,
// and after, application shutdown.
type AppCloser interface {
	PreShutdown()
	PostShutdown()
}

// CrashHandler defines the interface which should be implemented
// and registered via SetCrashHandler() in order to call
// custom recovery logic for otherwise unhandled application panics.
type CrashHandler interface {
	OnCrash(crashData interface{})
}

// LoopHandler defines the interface which should be implemented
// and registered via SetLoopHandler() to execute user code before
// and after the primary application channel loop. The primary loop runs
// after application initialization and is triggered by the application 
// heartbeat (if configured), or via manual calls to MsgPump().
type LoopHandler interface {
	PreLoop()
	PostLoop()
}

// Initialized returns a bool value denoting whether the application has
// been successfully initialized or not.
func Initialized() bool {
	return initialized
}

// MsgPump manually pulses the main application loop.
func MsgPump() {
	msgPump<- true
}

// Name returns the name of the GoApp application.
func Name() string {
	return appName
}

// SetAppStarter sets the AppStarter for the application. Must be done before
// Start() in order to matter.
func SetAppStarter(obj AppStarter) {
	mutex.Lock()
	defer mutex.Unlock()

	appStarter = obj
}

// SetAppCloser sets the AppCloser for the application.
func SetAppCloser(obj AppCloser) {
	mutex.Lock()
	defer mutex.Unlock()

	appCloser = obj
}

// SetCrashHandler sets the CrashHandler for the application.
func SetCrashHandler(obj CrashHandler) {
	mutex.Lock()
	defer mutex.Unlock()

	crashHandler = obj
}

// SetHeartbeat sets and, if appropriate, starts the heartbeat of the
// GoApp.
func SetHeartbeat(intervalMs int) {
	if intervalMs < 1 {
		syncObj.StopHeart()
		return
	}

	syncObj.StartHeart(intervalMs)
}

// SetLoopHandler sets the LoopHandler for the application.
func SetLoopHandler(obj LoopHandler) {
	mutex.Lock()
	defer mutex.Unlock()

	loopHandler = obj
}

// Start sets the GoApp's name and starts its execution.
func Start(name string) {
	startApp(name)

	// TODO: if service/daemon mode
	// go startApp(name)
}

// Stop begins the shutdown process of the application.
func Stop() {
	syncObj.Shutdown()
}

// handlePanic is called on the way out of the startApp function. It
// looks for unhandled panics and passes them to the registered CrashHandler.
func handlePanic() {
	err := recover()
	if err == nil {
		return 
	}

	crashHandler.OnCrash(err)
}

// internalInit performs initialization logic that can't be overridden
// by client code. The sequence of init operatiosn is AppStarter.PreInit(),
// internalInit, AppStarter.PostInit().
func internalInit() {
	log.Debug("App Init")
}

// internalShutdown performs clean-up logic that can't be overriden by
// client code. The sequence of shutdown operations is AppCloser.PreShutdown(),
// internalShutdown(), and then AppCloser.PostShutdown().
func internalShutdown() {
	log.Shutdown()
	log.Debug("App shutdown complete")
}

// startApp is the primary entry point for a GoApp. It performs basic application
// setup, runs the main application loop, and defers crash handling and shutdown
// operations until the function exits.
func startApp(name string) {
	defer internalStop()
	defer handlePanic()

	appName = name

	appStarter.PreInit()
	internalInit()
	appStarter.PostInit()

	initialized = true

	for syncObj.QueryRun() {
		loopHandler.PreLoop()

		select {
		case <-msgPump:
		case <-syncObj.QueryHeartbeat():
		case <-syncObj.QueryShutdown():
		}

		loopHandler.PostLoop()
	}
}

// internalStop marks the application as uninitialized, calls AppCloser and
// internal shutdown code, and then lets the lifecycle syncObj know that shutdown
// is complete.
func internalStop() {
	initialized = false
	syncObj.StopHeart()

	appCloser.PreShutdown()
	internalShutdown()
	appCloser.PostShutdown()

	syncObj.ShutdownComplete()
}

