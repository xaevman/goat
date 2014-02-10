//  ---------------------------------------------------------------------------
//
//  stack.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package diag

// External imports.
import (
	"github.com/xaevman/goat/core/log"
)

// Stdlib imports.
import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// Maximum number of stack records to hold.
const STACK_BUFFERS = 2045


// StackTrace represents the collection of stack frame data associated
// with a single goroutine.
type StackTrace struct {
	Frames []*StackFrame
}

// String pretty-prints a StackTrace object.
func (this *StackTrace) String() string {
	var buffer bytes.Buffer

	for i := range this.Frames {
		buffer.WriteString(this.Frames[i].String() + "\n")
	}

	return buffer.String()
}


// StackFrame represents the function name, file and line number information
// for a given frame within a stack trace.
type StackFrame struct {
	File  string
	Line  int
	Name  string
}

// String pretty-prints a StackFrame object.
func (this *StackFrame) String() string {
	return fmt.Sprintf(
		"%s :: %s:%d", 
		this.Name, 
		this.File, 
		this.Line,
	)
}


// NewStackTrace is a constructor function which queries the Go runtime
// for goroutine information and builds a StackTrace object for each.
// A slice of StackTraces, one for each running goroutine, is returned.
func NewStackTrace() []*StackTrace {
	stackRecords := make([]runtime.StackRecord, STACK_BUFFERS)
	count, ok    := runtime.GoroutineProfile(stackRecords)
	if !ok {
		log.Error(
			"Error creating stack trace. Buffer too small " +
			"(size:%d , needed:%d",
			STACK_BUFFERS,
			count,
		)

		return nil
	}

	results := make([]*StackTrace, 0)

	for i := 0; i < count; i++ {
		st       := new(StackTrace)
		st.Frames = make([]*StackFrame, 0)
		frames   := stackRecords[i].Stack()

		for x := range frames {
			f          := runtime.FuncForPC(frames[x])
			file, line := f.FileLine(frames[x])
			sf         := new(StackFrame)
			sf.Name     = strings.TrimSpace(filepath.Base(f.Name()))
			sf.File     = strings.TrimSpace(file)
			sf.Line     = line

			st.Frames = append(st.Frames, sf)
		}

		results = append(results, st)
	}

	return results
}
