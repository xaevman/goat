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

// Package dbg implements a protocol which can be included in a goat 
// project to enable realtime performance and diagnostic monitoring over 
// a TCP connection.
package dbg

// Products import.
import "github.com/xaevman/goat/prod"

// Stdlib imports.
import (
	"log"
	"testing"
)

// TestMsgSig tests to make sure that the message handler returns
// the expected message signature.
func TestMsgSig(t *testing.T) {
	handler := new(CmdMsgHandler)
	if handler.Signature() != prod.DBG_MSG {
		t.Fatalf(
			"Signature mismatch (%d vs %d)", 
			handler.Signature(), 
			prod.DBG_MSG,
		)
	}

	log.Println("TestMsgSig: passed")
}

// TestMsgSerialize tests serialization/deserialization of Msg objects.
func TestMsgSerialize(t *testing.T) {
	handler := new(CmdMsgHandler)
	cmd     := new(CmdMsg)
	cmd.Cmd  = CMD_JDUMP
	cmd.Data = ""

	b, err := handler.SerializeMsg(cmd)
	if err != nil {
		t.Fatal(err)
	}

	obj, err := handler.DeserializeMsg(b, 255)
	if err != nil {
		t.Fatal(err)
	}

	newCmd, ok := obj.(*CmdMsg)
	if !ok {
		t.Fatal("Invalid type received %T", obj)
	}

	if cmd.Cmd != newCmd.Cmd {
		t.Fatalf(
			"Cmd mismatch: %s vs %s", 
			cmd.Cmd, 
			newCmd.Cmd,
		)
	}

	if cmd.Data != newCmd.Data {
		t.Fatalf(
			"Data mismatch: %s vs %s", 
			cmd.Data, 
			newCmd.Data,
		)	}

	log.Printf("TestMsgSerialize: passed")
}
