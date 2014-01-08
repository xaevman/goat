//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

import(
	"math/rand"
	"net"
	"testing"
	"time"
)

var data []byte

func TestTCPSrv(t *testing.T) {
	data = make([]byte, 32 * 1024)

	fillData()

	srv := NewTCPSrv()
	go srv.Start("127.0.0.1:6600")

	<-time.After(1 * time.Second)

	for i := 0; i < 100; i++ {
		<-time.After(time.Duration(rand.Intn(200)) * time.Millisecond)
		go runClient(t)
	}

	<-time.After(15 * time.Second)
	srv.Stop()
}

func fillData() {
	for i := 0; i < len(data); i++ {
		data[i] = byte(rand.Intn(100))
	}
}

func runClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		conn.Write([]byte(data)[:rand.Intn(32 * 1024)])

		<-time.After(time.Duration(rand.Intn(15)) * time.Millisecond)
	}
}