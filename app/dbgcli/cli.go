//  ---------------------------------------------------------------------------
//
//  cli.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package main

// External imports.
import (
    "github.com/xaevman/goat/core/goapp"
    "github.com/xaevman/goat/core/log"
    "github.com/xaevman/goat/core/net"
    "github.com/xaevman/goat/lib/console"
    "github.com/xaevman/goat/lib/lifecycle"
    "github.com/xaevman/goat/prod"
    "github.com/xaevman/goat/prod/dbg"
)

// Stdlib imports
import (
    "errors"
    "fmt"
    "strings"
)

// Console text constants.
const (
    EXIT_MSG = "exit\n"
)

// Text styles.
var (
    errStyle = console.Style{
        ForeColor: console.FG_RED,
        Bold:      true,
    }

    sysStyle = console.Style{
        ForeColor: console.FG_YELLOW,
        Bold:      true,
    }

    txtStyle = console.Style{
        ForeColor: console.FG_WHITE,
    }

    privStyle = console.Style{
        ForeColor: console.FG_MAGENTA,
    }
)

// CmdInfo stores command, help text, and raw enum values for a given
// prod.DBG_MSG subtype.
type CmdInfo struct {
    cmdTxt  string
    helpTxt string
    val     byte
}

// Map of commands to info objects.
var cmdMap = map[string]*CmdInfo {
    "blocked" : &CmdInfo { "blocked", "Blocked goroutines"       , dbg.CMD_BLOCKED },
    "env"     : &CmdInfo { "env"    , "Environment variable data", dbg.CMD_ENV },
    "stack"   : &CmdInfo { "stack"  , "Full stack data",           dbg.CMD_STACK },
    "mem"     : &CmdInfo { "mem"    , "Memory allocation data",    dbg.CMD_MEM },
    "perf"    : &CmdInfo { "perf"   , "Performance counter data",  dbg.CMD_PERF },
    "sys"     : &CmdInfo { "sys"    , "General system data",       dbg.CMD_SYS },
}


// DbgCli represents a basic, command-line driven debugging client.
type DbgCli struct {
    inputSync  *lifecycle.Lifecycle
    msgHandler *dbg.CmdMsgHandler
    proto      *net.Protocol
    srvId      uint32
}
    
// Close deletes the CmdMsgHandler signature registration from the parent
// protocol, shuts down the console, and begins the application shutdown
// process.
func (this *DbgCli) Close() {
    this.proto.DeleteSignature(this.msgHandler)
    this.msgHandler = nil

    this.inputSync.Shutdown()
    console.Write(console.ENABLE_LINEWRAP)
    goapp.Stop()
}

// Init saves a reference to the parent protocol, registers the CmdMsgHandler
// signature on the protocol, and starts the console input goroutine.
func (this *DbgCli) Init(proto *net.Protocol) {
    this.inputSync  = lifecycle.New()
    this.msgHandler = new(dbg.CmdMsgHandler)
    this.proto      = proto

    this.proto.AddSignature(this.msgHandler)
    this.proto.SetAccessProvider(new(net.NoSecurity))

    go this.startInput()
}

// OnConnect saves a reference to the net ID of the newly connected server
// and prints welcome and help text to the console.
func (this *DbgCli) OnConnect(con net.Connection) {
    log.Debug("OnConnect %s", con.RemoteAddr())

    this.srvId = con.Id()

    console.ClearScreen()
    console.Write(console.DISABLE_LINEWRAP)
    console.WriteLine("====================================")
    console.WriteLine("DbgCli v0.1")
    console.WriteLine("Copyright 2014 Jared Chavez")
    console.WriteLine("====================================")
    console.WriteLine("")

    this.printHelp()
}

// OnDisconnect begins the protocol shutdown process.
func (this *DbgCli) OnDisconnect(con net.Connection) {
    log.Error("Disconnected from server")
    go goapp.Stop()
}

// OnError passes network error messages along to the logging service.
func (this *DbgCli) OnError(err error) {
    log.Error(err.Error())
}

// OnReceive makes sure that new incoming messages pass a type assertion
// and then routes the message to the appropriate command handler by
// sub-type.
func (this *DbgCli) OnReceive(msg interface{}, fromId uint32, access byte) {
    cmdMsg, ok := msg.(*dbg.CmdMsg)
    if !ok {
        log.Error("Cannot handle message type %T", cmdMsg)
        return
    }

    cmdMsg.FromId = fromId
    cmdMsg.Access = access

    switch cmdMsg.Cmd {
    default:
        this.printChatText(
            fmt.Sprintf(
                "Unknown cmd: %s %s", 
                cmdMsg.Cmd, 
                cmdMsg.Data,
            ),
            errStyle,
        )
        break
    case dbg.CMD_RESPONSE:
        this.printResponse(cmdMsg)
        break
    case dbg.CMD_ERROR:
        this.OnError(errors.New(cmdMsg.Data))
        break
    }
}

// OnShutdown begins the goapp shutdown process.
func (this *DbgCli) OnShutdown() {
    goapp.Stop()
}

// OnTimeout passes timeout events along to the logging system as an error.
// It does not make any attempts to retry failures.
func (this *DbgCli) OnTimeout(timeout *net.TimeoutEvent) {
    log.Error("Timeout: %v", timeout)
}



// handleInput is called when a new line of input text is received from
// the console. If the supplied text matches EXIT_MSG, the shutdown
// sequence is started. Otherwise, the text is parsed as either a help
// or debugging command.
func (this *DbgCli) handleInput(in string) {
    if in == EXIT_MSG {
        this.OnDisconnect(nil)
        return
    }

    in = strings.TrimSpace(in)
    if len(in) < 1 {
        return
    }

    console.Write(console.CURSOR_UP_ONE)
    console.Write(console.CLEAR_LINE)

    if in == "?" {
        this.printHelp()
        return
    }

    this.sendCmd(in)
}

// printChatText prints text to the console window in the specified
// style.
func (this *DbgCli) printChatText(txt string, style console.Style) {
    console.Write(console.CURSOR_UP_ONE)
    console.WriteLine("")
    console.WriteLineFmt(strings.TrimSuffix(txt, "\n"), style)
}

// printHelp prints the command help menu to the console.
func (this *DbgCli) printHelp() {
    this.printChatText("Command help", sysStyle)
    this.printChatText("?\t:\tPrint help text", sysStyle)
    this.printChatText("exit\t:\tExit DbgCli", sysStyle)

    for k, v := range cmdMap {
        this.printChatText(
            fmt.Sprintf("%s\t:\t%s", k, v.helpTxt),
            sysStyle,
        )
    }
}

// printResponse prints response messages to the console.
func (this *DbgCli) printResponse(cmdMsg *dbg.CmdMsg) {
    this.printChatText(cmdMsg.Data, txtStyle)
}

// sendCmd parses command line text into valid Dbg protocol commands,
// constructs CmdMsg objects and passes them along to the protocol layer
// for transmission.
func (this *DbgCli) sendCmd(cmd string) {
    if this.srvId == 0 {
        this.printChatText(
            fmt.Sprintf("Can't send, no socket"),
            errStyle,
        )
        return
    }

    parts := strings.SplitN(cmd, " ", 2)
    if len(parts) < 1 {
        this.printChatText(
            fmt.Sprintf("Invalid command format (<cmd> <args...>)", cmd),
            errStyle,
        )
        return
    }

    cmdInfo, exist := cmdMap[parts[0]]
    if !exist {
        this.printChatText(
            fmt.Sprintf("Unable to parse command %s", parts[0]),
            errStyle,
        )
        return
    }

    cmdMsg     := new(dbg.CmdMsg)
    cmdMsg.Cmd  = cmdInfo.val

    if len(parts) > 1 {
        cmdMsg.Data = parts[1]
    }

    this.printChatText(
        fmt.Sprintf("%s (%d)", cmdInfo.cmdTxt, cmdInfo.val), 
        txtStyle,
    )

    this.proto.SendMsg(this.srvId, prod.DBG_MSG, cmdMsg)
}

// startInput starts the console input loop, reading text from
// stdin.
func (this *DbgCli) startInput() {
    inChan := console.ReadInput(EXIT_MSG)

    for this.inputSync.QueryRun() {
        select {
        case in := <-inChan:
            this.handleInput(in)
        case <-this.inputSync.QueryShutdown():
        }
    }

    this.inputSync.ShutdownComplete()
}
