//  ---------------------------------------------------------------------------
//
//  chatcliapp.go
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
    "github.com/xaevman/goat/mod/goapp"
)

// Stdlib imports.
import (
    "flag"
)

// ConsoleCliStart is a goapp.AppStarter implementation which runs a
// ConsoleCli based ChatCli instance.
type ConsoleCliStart struct {}

// PreInit parses command line arguments to set the target server
// address and start-up username.
func (this *ConsoleCliStart) PreInit() {
    flag.StringVar(&srvAddr, "s", srvAddr, "remote server address")
    flag.StringVar(&chatCli.username, "u", userName, "chat username")
    flag.BoolVar(&useUdp, "udp", false, "enable UDP transport")
    flag.Parse()
}

// PostInit connects to the remote server, closing the application if
// a connection cannot be established.
func (this *ConsoleCliStart) PostInit() {
    var err error

    if useUdp {
        sock, err := chatproto.ListenUdp("127.0.0.1:8902")
        if err != nil {
            goapp.Stop()
            return
        }

        err = chatproto.DialUdp(srvAddr, sock)
    } else {
        err = chatproto.DialTcp(srvAddr)
    }

    if err != nil {
        goapp.Stop()
    }
}
