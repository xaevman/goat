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
	"github.com/xaevman/goat/prod/chat"
)

// Stdlib imports.
import (
	"os"
)

// Configurables.
var (
	srvAddr  = "127.0.0.1:8900"
	userName = "Anon"
)


// ConsoleCliStart goapp.AppStarter implementation which runs a 
// ConsoleCli based ChatCli instance.
type ConsoleCliStart struct {}

func (this *ConsoleCliStart) PreInit() {
	if len(os.Args) > 1 {
		srvAddr = os.Args[1]
	}
	if len(os.Args) > 2 {
		userName = os.Args[2]
	}
}
func (this *ConsoleCliStart) PostInit() {
	adapter := NewConsoleCli()
	chatCli := chat.NewChatCli(adapter)
	chatCli.SetUsername(userName)
	
	err := chatCli.Connect(srvAddr)
	if err != nil {
		goapp.Stop()
	}
}

// main is the application entry point.
func main() {
	goapp.SetAppStarter(new(ConsoleCliStart))

	stopChan := make(chan bool, 0)
	goapp.Start("ConsoleChatCli", stopChan)

	<-stopChan
}
