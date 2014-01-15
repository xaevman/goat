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

package lifecycle

import (
	"log"
	"testing"
	"time"
)

var syncObj *Lifecycle

func TestLifecycleWithHeartbeat(t *testing.T) {
	log.Println("TestLifecycleWithHeartbeat: startup")
	syncObj = New()

	log.Println("TestLifecycleWithHeartbeat: start heart")
	syncObj.StartHeart(1000)

	log.Println("TestLifecycleWithHeartbeat: doWork")
	go doWork()

	log.Println("TestLifecycleWithHeartbeat: wait 5 sec")
	<-time.After(5 * time.Second)

	log.Println("TestLifecycleWithHeartbeat: shutdown")
	syncObj.Shutdown()

	log.Println("TestLifecycleWithHeartbeat: passed")
}

func TestLifecycle(t *testing.T) {
	log.Println("TestLifecycleSync: startup")
	syncObj = New()

	log.Println("TestLifecycleSync: doWork")
	go doWork()

	log.Println("TestLifecycleSync: shutdown")
	syncObj.Shutdown()

	log.Println("TestLifecycleSync: passed")
}

func doWork() {
	for syncObj.QueryRun() {
		select {
		case <-syncObj.QueryHeartbeat():
			log.Println("Heartbeat received")
		case <-syncObj.QueryShutdown():
		}
	}

	log.Println("doWork: shutdown complete")
	syncObj.ShutdownComplete()
}
