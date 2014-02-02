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

package time

// Stdlib imports.
import (
	"sync"
	"time"
)

// Stopwatch presents a simple API for measuring the amount of time between
// two events.
type Stopwatch struct {
	duration  time.Duration
	mutex     sync.Mutex
	running   bool
	startTime time.Time
}

// Mark returns the elapsed time.Duration without stopping the underlying
// timer.
func (this *Stopwatch) Mark() time.Duration {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.running {
		this.duration = time.Since(this.startTime)
	}

	return this.duration
}

// MarkMs returns the elapsed time in whole milliseconds without stopping
// the underlying timer.
func (this *Stopwatch) MarkMs() int64 {
	return int64(this.Mark() / time.Millisecond)
}

// MarkSec returns the elapsed time in whole seconds without stopping the
// underlying timer.
func (this *Stopwatch) MarkSec() int64 {
	return int64(this.Mark() / time.Second)
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
	this.startTime = time.Now()
}

// Stop ends the time measurement, storing and returning the elapsed duration.
func (this *Stopwatch) Stop() time.Duration {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.duration = time.Since(this.startTime)
	this.running  = false

	return this.duration
}
