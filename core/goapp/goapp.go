//  ---------------------------------------------------------------------------
//
//  goapp.go
//
//  Copyright (c) 2014, Jared Chavez.
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
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/lib/time"
)

// Stdlib imports.
import (
	"os"
	"os/signal"
	"sync"
)

// Built in performance timers.
const (
	PERF_APP_TIMER_PRE_INIT = iota
	PERF_APP_TIMER_POST_INIT
	PERF_APP_TIMER_PRE_SHUTDOWN
	PERF_APP_TIMER_POST_SHUTDOWN
	PERF_APP_TIMER_RUNTIME
	PERF_APP_TIMER_PRE_LOOP
	PERF_APP_TIMER_POST_LOOP
	PERF_APP_TIMER_ON_HEARTBEAT
	PERF_APP_TIMER_LOOP_IDLE
	PERF_APP_MSGPUMP
	PERF_APP_COUNT
)

// Built in performance timer names.
var appTimerNames = []string {
	"TimerPreInitMs",
	"TimerPostInitMs",
	"TimerPreShutdownMs",
	"TimerPostShutdownMs",
	"TimerRuntime",
	"TimerPreLoopMs",
	"TimerPostLoopMs",
	"TimerOnHeartbeatMs",
	"TimerLoopIdleMs",
	"ManualMsgPump",
}

// Application properties.
var (
	appName     string
	appPerfs    = perf.NewCounterSet(
		"Service.GoApp", 
		PERF_APP_COUNT, 
		appTimerNames,
	)
	exitCode    = 0
	initialized = false
	runTimer    = new(time.Stopwatch)
	stopwatch   = new(time.Stopwatch)
)

// Synchronization helpers.
var (
	msgPump  = make(chan bool, 1)
	mutex    sync.Mutex
	stopChan chan bool
	syncObj  = lifecycle.New()
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
	OnHeartbeat()
	PreLoop()
	PostLoop()
}

// MsgPump manually pulses the main application loop.
func MsgPump() {
	msgPump <- true
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

// SetExitCode sets the exit code the application should return
// when shutdown is complete.
func SetExitCode(code int) {
	exitCode = code
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
func Start(name string, callback chan bool) {
	runTimer.Start()

	stopChan = callback
	go startApp(name)
}

// Stop begins the shutdown process of the application.
func Stop() {
	appPerfs.Set(PERF_APP_TIMER_RUNTIME, int64(runTimer.MarkSec()))

	go func() {
		syncObj.Shutdown()
	}()
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

	// enable stats on all relevant counters
	appPerfs.EnableStats(PERF_APP_TIMER_LOOP_IDLE)
	appPerfs.EnableStats(PERF_APP_TIMER_ON_HEARTBEAT)
	appPerfs.EnableStats(PERF_APP_TIMER_POST_LOOP)
	appPerfs.EnableStats(PERF_APP_TIMER_PRE_LOOP)

	c := make(chan os.Signal, 0)
	signal.Notify(c, os.Interrupt)
	go func(){
		for {
		    select {
		    case <-c:
	    		Stop()
		    	return
	    	}
    	}
	}()
}

// internalShutdown performs clean-up logic that can't be overriden by
// client code. The sequence of shutdown operations is AppCloser.PreShutdown(),
// internalShutdown(), and then AppCloser.PostShutdown().
func internalShutdown() {
	log.Shutdown()
	log.Debug("**** Perf snapshot ***\n%v", perf.TakeSnapshot())
	log.Debug("App shutdown complete")
}

// startApp is the primary entry point for a GoApp. It performs basic application
// setup, runs the main application loop, and defers crash handling and shutdown
// operations until the function exits.
func startApp(name string) {
	defer internalStop()
	defer handlePanic()

	appName = name

	stopwatch.Restart()
	appStarter.PreInit()
	appPerfs.Set(PERF_APP_TIMER_PRE_INIT, stopwatch.MarkMs())

	internalInit()

	stopwatch.Restart()
	appStarter.PostInit()
	appPerfs.Set(PERF_APP_TIMER_POST_INIT, stopwatch.MarkMs())

	stopwatch.Reset()

	for syncObj.QueryRun() {
		stopwatch.Restart()
		loopHandler.PreLoop()
		appPerfs.Set(PERF_APP_TIMER_PRE_LOOP, stopwatch.MarkMs())

		select {
		case <-msgPump:
			appPerfs.Set(PERF_APP_TIMER_LOOP_IDLE, stopwatch.MarkMs())
			appPerfs.Increment(PERF_APP_MSGPUMP)
		case <-syncObj.QueryHeartbeat():
			appPerfs.Set(PERF_APP_TIMER_LOOP_IDLE, stopwatch.MarkMs())
			
			stopwatch.Restart()
			loopHandler.OnHeartbeat()
			appPerfs.Set(PERF_APP_TIMER_ON_HEARTBEAT, stopwatch.MarkMs())
		case <-syncObj.QueryShutdown():
			appPerfs.Set(PERF_APP_TIMER_LOOP_IDLE, stopwatch.MarkMs())
		}

		stopwatch.Restart()
		loopHandler.PostLoop()
		appPerfs.Set(PERF_APP_TIMER_POST_LOOP, stopwatch.MarkMs())

		stopwatch.Restart()
	}
}

// internalStop marks the application as uninitialized, calls AppCloser and
// internal shutdown code, and then lets the lifecycle syncObj know that shutdown
// is complete.
func internalStop() {
	syncObj.StopHeart()

	stopwatch.Restart()
	appCloser.PreShutdown()
	appPerfs.Set(PERF_APP_TIMER_PRE_SHUTDOWN, stopwatch.MarkMs())

	internalShutdown()

	stopwatch.Restart()
	appCloser.PostShutdown()
	appPerfs.Set(PERF_APP_TIMER_POST_SHUTDOWN, stopwatch.MarkMs())

	syncObj.ShutdownComplete()

	os.Exit(exitCode)
}
