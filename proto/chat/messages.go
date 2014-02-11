//  ---------------------------------------------------------------------------
//
//  messages.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package chat

import _ "github.com/xaevman/goat/proto"

/* +NetMsg+ proto.CHAT_MSG */
// Msg represents a message sent from one client to the chat system.
// ChannelId is the channel that the message is being sent to. From is the 
// friendly name of the user sending the message. FromId is the NetID of
// the message sender. Subtype denotes the type of chat message being sent.
// ToID is the NetID of the message recipient. Text is the actual text of 
// the message.
type Msg struct {
    Access    byte
    ChannelId uint32    // +export+
    From      string    // +export+
    FromId    uint32
    Subtype   byte      // +export+
    Text      string    // +export+
}
