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

// TestChatMsg tests serialization/deserialization of ChatMsg objects.
func TestChatMsgSerialize(t *testing.T) {
	cm := ChatMsg {
		ChannelId: 10,
		From:      "Jared",
		FromId:    12345,
		ToId:      54321,
		Text:      "This is my test message! Rawr",
	}

	b     := SerializeChatMsg(&cm)
	newCm := DeserializeChatMsg(b)

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

	log.Printf("TestChatMsgSerialize: passed")
}
