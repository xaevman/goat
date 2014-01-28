//  ---------------------------------------------------------------------------
//
//  time.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package time implements helper functions and objects to assist in
// measuring time-based tasks.
package time

// Stdlib imports.
import (
	"sync"
	stdtime "time"
)

// Stopwatch presents a simple API for measuring the amount of time between
// two events.
type Stopwatch struct {
	duration  stdtime.Duration
	mutex     sync.Mutex
	running   bool
	startTime stdtime.Time
}

// Mark returns the elapsed time.Duration without stopping the underlying
// timer.
func (this *Stopwatch) Mark() stdtime.Duration {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.running {
		this.duration = stdtime.Since(this.startTime)
	}

	return this.duration
}

// MarkMs returns the elapsed time in whole milliseconds without stopping
// the underlying timer.
func (this *Stopwatch) MarkMs() int64 {
	return int64(this.Mark() / stdtime.Millisecond)
}

// MarkSec returns the elapsed time in whole seconds without stopping the
// underlying timer.
func (this *Stopwatch) MarkSec() int64 {
	return int64(this.Mark() / stdtime.Second)
}

// Restart resets and then starts the current timer.
func (this *Stopwatch) Restart() {
	this.Reset()
	this.Start()
}

// Reset stops the underlying timer and resets the elapsed duration to zero.
func (this *Stopwatch) Reset() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.duration = 0
	this.running  = false
}

// Start marks the starting point for measurement, effectively starting the
// watch.
func (this *Stopwatch) Start() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.running   = true
	this.startTime = stdtime.Now()
}

// Stop ends the time measurement, storing and returning the elapsed duration.
func (this *Stopwatch) Stop() stdtime.Duration {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.duration = stdtime.Since(this.startTime)
	this.running  = false

	return this.duration
}
