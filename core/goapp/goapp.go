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

package goapp


import (
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/lib/lifecycle"
)


type AppStarter interface {
	PreInit()
	PostInit()
}


type AppCloser interface {
	PreShutdown()
	PostShutdown()
}


type CrashHandler interface {
	OnCrash(crashData interface{})log.Shutdown
}


type LoopHandler interface {
	PreLoop()
	OnHeartbeat()
	PostLoop()
}


type GoApp struct {
	appStarter   AppStarter
	appCloser    AppCloser
	crashHandler CrashHandler
	loopHandler  LoopHandler
	name         string
	syncObj      *lifecycle.Lifecycle
}


func (this *GoApp) Name() string {
	return this.name
}


func (this *GoApp) Start(name string, syncObj *lifecycle.Lifecycle) {
	defer this.startShutdown()
	defer this.handlePanic()

	this.name    = name
	this.syncObj = syncObj

	this.appStarter.PreInit()
	this.internalInit()
	this.appStarter.PostInit()

	for this.syncObj.QueryRun() {
		this.loopHandler.PreLoop()

		select {
		case <-this.syncObj.QueryHeartbeat():
		case <-this.syncObj.QueryShutdown():
		}

		this.loopHandler.PostLoop()
	}
}


func (this *GoApp) handlePanic() {
	err := recover()
	if err == nil {
		return 
	}

	this.crashHandler.OnCrash(err)
}


func (this *GoApp) internalInit() {
	log.Debug("App Init")
}


func (this *GoApp) internalShutdown() {
	log.Shutdown()
	log.Debug("App shutdown complete")
}


func (this *GoApp) startShutdown() {
	this.appCloser.PreShutdown()
	this.internalShutdown()
	this.appCloser.PostShutdown()

	this.syncObj.ShutdownComplete()
}

