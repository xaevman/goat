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

package dbg

// External imports.
import (
	"github.com/xaevman/goat/core/goapp"
	"github.com/xaevman/goat/core/log"
	"github.com/xaevman/goat/core/net"
	"github.com/xaevman/goat/lib/console"
	"github.com/xaevman/goat/lib/lifecycle"
	"github.com/xaevman/goat/prod"
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

type CmdInfo struct {
	cmdTxt  string
	helpTxt string
	val     byte
}

var cmdMap = map[string]*CmdInfo {
	"env"   : &CmdInfo { "env"   , "Environment variable data", CMD_ENV },
	"stack" : &CmdInfo { "stack" , "Full stack data",           CMD_STACK },
	"mem"   : &CmdInfo { "mem"   , "Memory allocation data",    CMD_MEM },
	"perf"  : &CmdInfo { "perf"  , "Performance counter data",  CMD_PERF },
	"sys"   : &CmdInfo { "sys"   , "General system data",       CMD_SYS },
}


//
type DbgCli struct {
	inputSync  *lifecycle.Lifecycle
	msgHandler *CmdMsgHandler
	proto      *net.Protocol
	srvId      uint32
}
	
// Close deletes the CmdMsgHandler signature registration from the parent
// protocol.
func (this *DbgCli) Close() {
	this.proto.DeleteSignature(this.msgHandler)
	this.msgHandler = nil

	this.inputSync.Shutdown()
	console.Write(console.ENABLE_LINEWRAP)
	goapp.Stop()
}

// Init saves a reference to the parent protocol and registers the CmdMsgHandler
// signature on the protocol.
func (this *DbgCli) Init(proto *net.Protocol) {
	this.inputSync  = lifecycle.New()
	this.msgHandler = new(CmdMsgHandler)
	this.proto      = proto

	this.proto.AddSignature(this.msgHandler)
	this.proto.SetAccessProvider(new(net.NoSecurity))

	go this.startInput()
}

// OnConnect logs debugging information about the newly connected client.
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

// OnDisconnect logs debugging information about the newly disconnected
// client.
func (this *DbgCli) OnDisconnect(con net.Connection) {
	this.proto.Shutdown()
	log.Error("Disconnected from server")
}

// OnError passes network error messages along to the logging service.
func (this *DbgCli) OnError(err error) {
	log.Error(err.Error())
}

// OnReceive makes sure that new incoming messages pass a type assertion
// and then routes the message to the appropriate command handler.
func (this *DbgCli) OnReceive(msg interface{}) {
	cmdMsg, ok := msg.(*CmdMsg)
	if !ok {
		log.Error("Cannot handle message type %T", cmdMsg)
		return
	}

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
	case CMD_RESPONSE:
		this.printResponse(cmdMsg)
		break
	case CMD_ERROR:
		this.OnError(errors.New(cmdMsg.Data))
		break
	}
}

// OnShutdown performs no actions in DbgSrv.
func (this *DbgCli) OnShutdown() {
	goapp.Stop()
}

// OnTimeout passes timeout events along to the logging system as an error.
// It does not make any attempts to retry failures.
func (this *DbgCli) OnTimeout(timeout *net.TimeoutEvent) {
	log.Error("Timeout: %v", timeout)
}


func (this *DbgCli) printResponse(cmdMsg *CmdMsg) {
	this.printChatText(cmdMsg.Data, txtStyle)
}


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

	cmdMsg     := new(CmdMsg)
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


// printChatText prints any text which should go into the chat window.
// The lone exception is text input via Stdin, which should be handled by
// printTextFromInput because of the extra new line that will be input
// from that method.
func (this *DbgCli) printChatText(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.WriteLine("")
	console.WriteLineFmt(strings.TrimSuffix(txt, "\n"), style)
}


// printTextFromInput handles printing the given text after accounting
// for the extra new line that will be present from text input through
// Stdin.
func (this *DbgCli) printTextFromInput(txt string, style console.Style) {
	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.WriteLineFmt(txt, style)
}

// printPrompt inserts the given number of empty lines and then outputs the
// prompt.
func (this *DbgCli) printPrompt(spacing int) {
	for i := 0; i < spacing; i++ {
		console.WriteLine("")
	}

	console.Write(console.CURSOR_UP_ONE)
	console.Write(console.CLEAR_LINE)
	console.SetBold()
	console.Write("> ")
	console.ClearFormat()
}


// handleInput is called when a new line of input text is received from
// the console. If the supplied text matches EXIT_MSG, the shutdown
// sequence is started. Otherwise, the text is parsed as either a command
// (if starting with /) or a chat message and sent to the server
// appropriately.
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
