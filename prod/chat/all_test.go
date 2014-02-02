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

// External imports.
import (

)

// Stdlib imports.
import (
	"log"
	"testing"
)

// TestMsg tests serialization/deserialization of Msg objects.
func TestMsgSerialize(t *testing.T) {
	handler := new(MsgHandler)
	cm      := Msg {
		ChannelId: 10,
		From:      "Jared",
		FromId:    12345,
		Subtype:   MSG_SUB_CHAT,
		ToId:      54321,
		Text:      "This is my test message! Rawr",
	}

	b, err := handler.SerializeMsg(&cm)
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

	if cm.FromId != newCm.FromId {
		t.Fatalf(
			"From Ids do not match: %v != %v\n",
			cm.FromId,
			newCm.FromId,
		)
	}

	if cm.Subtype != newCm.Subtype {
		t.Fatalf(
			"Subtype mismatch: %v != %v\n",
			cm.Subtype,
			newCm.Subtype,
		)
	}

	if cm.ToId != newCm.ToId {
		t.Fatalf(
			"To Ids do not match: %v != %v\n",
			cm.ToId,
			newCm.ToId,
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
