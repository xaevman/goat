//  ---------------------------------------------------------------------------
//
//  chat.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package chat

// Chat module name.
const CHAT_MOD_NAME = "Chat"

// Timeout value for sends.
const CHAT_MSG_SEND_TIMEOUT_SEC = 5

// Msg subtypes
const (
    MSG_SUB_CHAT = iota
    MSG_SUB_CMD
    MSG_SUB_CONNECT
    MSG_SUB_JOIN_CHANNEL
    MSG_SUB_LEAVE_CHANNEL
    MSG_SUB_SET_NAME
)

// Channel constants.
const (
    PUB_CHANNEL = "public"
)
