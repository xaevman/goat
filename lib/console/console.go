//  ---------------------------------------------------------------------------
//
//  console.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package console provides some VT100-like helper functions for providing
// a more full-featured console presentation.
package console

// Stdlib imports.
import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// ANSI set mode flags.
const (
	S_CLEAR		= iota
	S_BOLD
	FG_BLACK 	= 28 + iota
	FG_RED
	FG_GREEN
	FG_YELLOW
	FG_BLUE
	FG_MAGENTA
	FG_CYAN
	FG_WHITE
	BG_BLACK 	= 30 + iota
	BG_RED
	BG_GREEN
	BG_YELLOW
	BG_BLUE
	BG_MAGENTA
	BG_CYAN
	BG_WHITE
)

// Config options.
const (
	READ_BUFFER_LEN = 20
)

// Key sequences.
var (
	CLEAR_FORMAT    = ESC_CHAR + "[0m"
	CLEAR_LINE      = ESC_CHAR + "[2K"
	CLEAR_SCREEN    = ESC_CHAR + "[2J"
	CURSOR_DOWN_ONE = ESC_CHAR + "[1B"
	CURSOR_UP_ONE   = ESC_CHAR + "[1A"
	ESC_CHAR        = string(0x1B)
	SCROLL_DOWN     = ESC_CHAR + "M"
	SCROLL_UP       = ESC_CHAR + "D"
	SET_BOLD        = ESC_CHAR + "[1m"
)

// Synchronization helpers.
var mutex sync.Mutex

// Style represents a console output format used by 
// WriteFmt() and WriteLineFmt().
type Style struct {
	ForeColor int
	BackColor int
	Bold      bool
}

// ClearFormat sets console formatting back to default.
func ClearFormat() {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Fprint(os.Stdout, CLEAR_FORMAT)
}

// ClearScreen clears the console screen.
func ClearScreen() {
	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprint(os.Stdout, CLEAR_SCREEN)
}

// ReadInput starts a go routine which reads input until exitSeq is observed,
// and returns a read-only channel on which the client can receive strings
// that are read from Stdin.
func ReadInput(exitSeq string) <-chan string {
	input    := bufio.NewReader(os.Stdin)
	readChan := make(chan string, READ_BUFFER_LEN)

	go func() {
	    for {
	    	txt, err := input.ReadString('\n')
	    	if err != nil {
	    		continue
	    	}

	    	if txt == exitSeq {
		    	readChan<- txt
	    		return
	    	}

	    	readChan<- txt
		}
	}()

	return readChan
}

// SetBackColor sets the background color for the console, starting at the
// current cursor position.
func SetBackColor(flag int) {
	if flag < BG_BLACK || flag > BG_WHITE {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprintf(
		os.Stdout,
		"%s[%vm",
		ESC_CHAR,
		flag,
	)
}

// SetBold sets the bold flag starting at the current cursor position.
func SetBold() {
	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprint(os.Stdout, SET_BOLD)
}

// SetForeColor sets the foreground color for the console, starting at the
// current cursor position.
func SetForeColor(flag int) {
	if flag < FG_BLACK || flag > FG_WHITE {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprintf(
		os.Stdout,
		"%s[%vm",
		ESC_CHAR,
		flag,
	)
}

// Write writes the supplied text, as is, to the console.
func Write(text string) {
	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprint(os.Stdout, text)
}

// WriteFmt sets the foreground, background color and font weight,
// specified by the given style, writes the given text, and then clears
// the formatting.
func WriteFmt(text string, style Style) {
	SetForeColor(style.ForeColor)
	SetBackColor(style.BackColor)
	
	if style.Bold {
		SetBold()
	}
	
	Write(text)
	
	ClearFormat()
}

// WriteLine writes the supplied text, terminated with a newline, to 
// the console.
func WriteLine(text string) {
	mutex.Lock()
	defer mutex.Unlock()
	
	fmt.Fprint(os.Stdout, text + "\n")
}

// WriteLineFmt sets sets foreground, background color and font weight,
// specified by the given style, writes the given text, and then clears 
// the formatting before writing out a newline. Doing so in this way 
// avoids color artifacts in the console.
func WriteLineFmt(text string, style Style) {
	WriteFmt(text, style)	
	Write("\n")
}

