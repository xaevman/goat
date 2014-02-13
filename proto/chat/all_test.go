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

package chat

// Products import.
import "github.com/xaevman/goat/proto"

// Stdlib imports.
import (
    "log"
    "testing"
)

// TestMsgSig tests to make sure that the message handler returns
// the expected message signature.
func TestMsgSig(t *testing.T) {
    handler := new(MsgHandler)
    if handler.Signature() != proto.CHAT_MSG {
        t.Fatalf(
            "Signature mismatch (%d vs %d)", 
            handler.Signature(), 
            proto.CHAT_MSG,
        )
    }

    log.Println("TestMsgSig: passed")
}

// TestMsgSerialize tests serialization/deserialization of Msg objects.
func TestMsgSerialize(t *testing.T) {
    handler     := new(MsgHandler)
    cm          := new(Msg)
    cm.ChannelId = 10
    cm.From      = "Jared"
    cm.Subtype   = MSG_SUB_CHAT
    cm.Text      = "This is my test message! Rawr"

    b, err := handler.SerializeMsg(cm)
    if err != nil {
        t.Fatal(err)
    }

    obj, err := handler.DeserializeMsg(b, 255)
    if err != nil {
        t.Fatal(err)
    }

    newCm, ok := obj.(*Msg)
    if !ok {
        t.Fatal("Invalid type received %T", obj)
    }

    if cm.ChannelId != newCm.ChannelId {
        t.Fatalf(
            "Channel Ids do not match: %v != %v\n",
            cm.ChannelId,
            newCm.ChannelId,
        )
    }

    if cm.From != newCm.From {
        t.Fatalf(
            "From text does not match: %v != %v\n",
            cm.From,
            newCm.From,
        )
    }

    if cm.Subtype != newCm.Subtype {
        t.Fatalf(
            "Subtype mismatch: %v != %v\n",
            cm.Subtype,
            newCm.Subtype,
        )
    }

    if cm.Text != newCm.Text {
        t.Fatalf(
            "Text does not match: %v != %v\n",
            cm.Text,
            newCm.Text,
        )
    }

    log.Printf("TestMsgSerialize: passed")
}
