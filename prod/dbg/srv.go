//  ---------------------------------------------------------------------------
//
//  srv.go
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
    "github.com/xaevman/goat/core/diag"
    "github.com/xaevman/goat/core/log"
    "github.com/xaevman/goat/core/net"
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/prod"
)

// Stdlib imports
import (
    "fmt"
)


// DbgSrv represents a basic debugging server. Attach this event handler
// to an existing or new protocol to gain some simple debugging capabilities
// over TCP connections from client implementations such as DbgCli.
type DbgSrv struct {
    msgHandler *CmdMsgHandler
    proto      *net.Protocol
}
    
// Close deletes the CmdMsgHandler signature registration from the parent
// protocol.
func (this *DbgSrv) Close() {
    this.proto.DeleteSignature(this.msgHandler)
    this.msgHandler = nil
}

// Init saves a reference to the parent protocol and registers the CmdMsgHandler
// signature on the protocol.
func (this *DbgSrv) Init(proto *net.Protocol) {
    this.msgHandler = new(CmdMsgHandler)
    this.proto      = proto

    this.proto.AddSignature(this.msgHandler)
    this.proto.SetAccessProvider(new(net.NoSecurity))
}

// OnConnect logs debugging information about the newly connected client.
func (this *DbgSrv) OnConnect(con net.Connection) {
    log.Debug("OnConnect %s", con.RemoteAddr())
}

// OnDisconnect logs debugging information about the newly disconnected
// client.
func (this *DbgSrv) OnDisconnect(con net.Connection) {
    log.Debug("OnDisconnect %s", con.RemoteAddr())
}

// OnError passes network error messages along to the logging service.
func (this *DbgSrv) OnError(err error) {
    log.Error(err.Error())
}

// OnReceive makes sure that new incoming messages pass a type assertion
// and then routes the message to the appropriate command handler.
func (this *DbgSrv) OnReceive(msg interface{}, fromId uint32, access byte) {
    cmdMsg, ok := msg.(*CmdMsg)
    if !ok {
        log.Error("Cannot handle message type %T", cmdMsg)
        return
    }

    cmdMsg.FromId = fromId
    cmdMsg.Access = access

    switch cmdMsg.Cmd {
    default:
        log.Error(
            "Unknown cmd: %s %s", 
            cmdMsg.Cmd, 
            cmdMsg.Data,
        )
        break
    case CMD_BLOCKED:
        this.onBlockedCmd(cmdMsg)
    case CMD_ENV:
        this.onEnvCmd(cmdMsg)
    case CMD_ERROR:
        log.Error(cmdMsg.Data)
    case CMD_STACK:
        this.onStackCmd(cmdMsg)
    case CMD_MEM:
        this.onMemCmd(cmdMsg)
    case CMD_PERF:
        this.onPerfCmd(cmdMsg)
    case CMD_RESPONSE:
        break
    case CMD_SYS:
        this.onSysCmd(cmdMsg)
    }
}

// OnShutdown performs no actions in DbgSrv.
func (this *DbgSrv) OnShutdown() {}

// OnTimeout passes timeout events along to the logging system as an error.
// It does not make any attempts to retry failures.
func (this *DbgSrv) OnTimeout(timeout *net.TimeoutEvent) {
    log.Error("Timeout: %v", timeout)
}


// onBlockedCmd dumps stack trace information for currently blocked 
// goroutines and transmits them back to the requestor.
func (this *DbgSrv) onBlockedCmd(cmdMsg *CmdMsg) {
    data       := diag.NewBlockedData()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = data

    this.send(cmdMsg)
}

// onEnvCmd dumps environment variable data and transmits it back to the
// requestor.
func (this *DbgSrv) onEnvCmd(cmdMsg *CmdMsg) {
    env        := diag.NewEnvData()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = fmt.Sprintf("%v", env)

    this.send(cmdMsg)
}

// onMemCmd dumps memory statistics data adn transmits it back to the
// requestor.
func (this *DbgSrv) onMemCmd(cmdMsg *CmdMsg) {
    mem        := diag.NewMemData()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = diag.FmtMemStatsStr(mem)

    this.send(cmdMsg)
}

// onPerfCmd dumps performance counter information and transmits it back
// to the requestor.
func (this *DbgSrv) onPerfCmd(cmdMsg *CmdMsg) {
    perfs      := perf.TakeSnapshot()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = perfs.StringBrief()

    this.send(cmdMsg)
}

// onStackCmd dumps a full stack trace of all goroutines and transmits it
// back to the requestor.
func (this *DbgSrv) onStackCmd(cmdMsg *CmdMsg) {
    stack      := diag.NewFullStackTrace()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = stack

    this.send(cmdMsg)
}

// onSysCmd gathers basic system information and transmits it back to
// the requestor.
func (this *DbgSrv) onSysCmd(cmdMsg *CmdMsg) {
    sys        := diag.NewSysData()
    cmdMsg.Cmd  = CMD_RESPONSE
    cmdMsg.Data = sys.String()

    this.send(cmdMsg)
}

// send passes the give CmdMsg along to the protocol layer.
func (this *DbgSrv) send(cmdMsg *CmdMsg) {
    this.proto.SendMsg(cmdMsg.FromId, prod.DBG_MSG, cmdMsg)
}
