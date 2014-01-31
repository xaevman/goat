//  ---------------------------------------------------------------------------
//
//  main.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package main

// External imports.
import (
	"github.com/xaevman/goat/core/goapp"
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/prod/chat"
)

// Stdlib imports.
import (
	"os"
	"strconv"
)

// Configurable settings.
var (
	maxTests = 100
	myIndex  = 0
	myName   = "Anon"
	srvAddr  = "127.0.0.1:8900"
)

// ChatCli notifier.
var adapter *ChatTest


// ChatTestStarter is a goapp.AppStarter implementation which sends
// messages and makes sure they are echo'd appropriately for stress
// testing ChatSrv. 
type ChatTestStarter struct {}

func (this *ChatTestStarter) PreInit() {
	if len(os.Args) > 1 {
		srvAddr = os.Args[1]
	}
	if len(os.Args) > 2 {
		myIndex, _ = strconv.Atoi(os.Args[2])
		myName     = "chattest" + os.Args[2]
	}
	if len(os.Args) > 3 {
		maxTests, _ = strconv.Atoi(os.Args[3])
	}

	log.Info(
		"Startup params (addr: %s, index: %d, tests: %d)",
		srvAddr,
		myIndex,
		maxTests,
	)
}
func (this *ChatTestStarter) PostInit() {
	// connect
	adapter  = NewChatTest(myIndex)
	chatCli := chat.NewChatCli(adapter)
	chatCli.SetUsername(myName)

	err := chatCli.Connect(srvAddr)
	if err != nil {
		goapp.Stop()
	}
}


// ChatTestCloser is a goapp.AppCloser implementation which prints
// the results of this client's part of the stress test.
type ChatTestCloser struct {}

func (this *ChatTestCloser) PreShutdown() {}
func (this *ChatTestCloser) PostShutdown() {
	success, errors, exeTime := adapter.GetResults()

	log.Info(
		"TEST RESULTS (Success: %d, Failure: %d)",
		success,
		errors,
	)
	log.Info(
		"TEST RESULTS (Execution time: %v)",
		exeTime,
	)
	log.Info(
		"TEST RESULTS (%.2f msg/sec)",
		float64(maxTests) / exeTime.Seconds(),
	)

	if success == maxTests {
		goapp.SetExitCode(0)
	}

	goapp.SetExitCode(errors)	
}


// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ChatTestStarter))
	goapp.SetAppCloser(new(ChatTestCloser))

	stopChan := make(chan bool, 0)
	goapp.Start("ChatTestCli", stopChan)

	<-stopChan
}
