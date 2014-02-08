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

package dbg

import _ "github.com/xaevman/goat/prod"

/* +NetMsg+ prod.DBG_MSG */
// CmdMsg represents a command message being sent into the debugging
// system. Access is the authorized level of access passed up from the
// protocol layer. Cmd is the base command being issued to the server.
// Data contains any arguments passed along with the base command.
// CmdMsg objects are re-used for the replies from client to server as
// well. In a reply, the Access field will contain the access level 
// returned from the server. If Access == 0, then an error occured and
// the relevant error message will be contained in the Data field. If
// there was no error the Data field will contain the JSON encoded
// debugging data from the server.
type CmdMsg struct {
	Access byte
	Cmd    byte		// +export+
	Data   string	// +export+
	FromId uint32
}
