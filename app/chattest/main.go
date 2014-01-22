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
	"github.com/xaevman/goat/lib/fs"
	"github.com/xaevman/goat/proto/chat"
)

// Stdlib imports.
import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Configurable settings.
var (
	maxTests = 100
	myIndex  = 0
	myName   = "Anon"
	srvAddr  = "127.0.0.1:8900"
)
// Test counters.
var (
	errors     = 0
	success    = 0
	startTime  time.Time
	testCount  = 0
	testCursor = 0
	totalIndex = 0
)	

// Current channel.
var currentChan uint32 = 0

// Chat text list.
var myPhrases = make([]string, 0)


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
		myName     = os.Args[2]
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
	// find my sentences
	file, _ := fs.OpenFile("./sentences.data")
	reader  := bufio.NewReader(file)
	for 
		line, err := reader.ReadString('.')
		err != io.EOF
		line, err = reader.ReadString('.') {

		line = strings.TrimSpace(line)
		myPhrases = append(myPhrases, line) 
	}

	log.Info("Total phrases: %d", len(myPhrases))

	// connect
	adapter := new(ChatTest)
	chatCli := chat.NewChatCli(adapter)
	chatCli.SetUsername(myName)

	startTime = time.Now()
	
	err := chatCli.Connect(srvAddr)
	if err != nil {
		goapp.Stop()
	}
}


// ChatTestCloser is a goapp.AppCloser implementation which prints
// the results of this client's part of the stress test.
type ChatTestCloser struct {}

func (this *ChatTestCloser) PreShutdown() {
	exeTime := time.Since(startTime)

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
}
func (this *ChatTestCloser) PostShutdown() {
	if success == len(myPhrases) {
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
