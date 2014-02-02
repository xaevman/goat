//  ---------------------------------------------------------------------------
//
//  dbg.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package dbg

// External imports.
import (
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/buffer"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/prod"
)

// Stdlib imports.
import (
	"errors"
	"fmt"
)

// Perf counters.
const (
	PERF_DBG_ERR_BAD_OBJ = iota
	PERF_DBG_ERR_DESERIALIZE
	PERF_DBG_RCV
	PERF_DBG_SEND
	PERF_DBG_COUNT
)

// Perf counter friendly names
var perfDbgNames = []string {
	"ErrorBadObject",
	"ErrorDeserializeFailed",
	"MessageReceived",
	"MessageSent",
}

// Perf counters.
var dbgPerfs = perf.NewCounterSet(
	"Debug",
	PERF_DBG_COUNT,
	perfDbgNames,
)


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
	Cmd    string
	Data   string
}


// CmdMsgHandler is a net.MsgProcessor implementation which handles 
// serialization for CmdMsg objects.
type CmdMsgHandler struct {}

// Close performs no actions in CmdMsgHandler.
func (this *CmdMsgHandler) Close() {}

// Init performs no actions in CmdMsgHandler.
func (this *CmdMsgHandler) Init(proto *net.Protocol) {}

// DeserializeMsg is called by the parent protocol when raw message data is
// received from the network module.
func (this *CmdMsgHandler) DeserializeMsg(msg *net.Msg, access byte) (interface{}, error)  {
	var err error

	cursor := 0
	data   := msg.GetPayload()
	cmdmsg := new(CmdMsg)

	cmdmsg.Cmd, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil, err }

	cmdmsg.Data, err = buffer.ReadString(data, &cursor)
	if err != nil { return nil, err }

	cmdmsg.Access = access

	return cmdmsg, nil
}

// SerializeMsg serializes a CmdMsg object.
func (this *CmdMsgHandler) SerializeMsg(data interface{}) (*net.Msg, error) {
	cursor     := 0
	cmdmsg, ok := data.(*CmdMsg)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot serialize type %T", data))
	}

	dataLen := 
		buffer.LenString(cmdmsg.Cmd) +
		buffer.LenString(cmdmsg.Data)

	dataBuffer := make([]byte, dataLen)

	buffer.WriteString(cmdmsg.Cmd, dataBuffer, &cursor)
	buffer.WriteString(cmdmsg.Data, dataBuffer, &cursor)

	msg := net.NewMsg()
	msg.SetMsgType(this.Signature())
	msg.SetPayload(dataBuffer)

	return msg, nil
}

func (this *CmdMsgHandler) Signature() uint16 {
	return prod.DBG_MSG
}
