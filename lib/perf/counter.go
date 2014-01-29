//  ---------------------------------------------------------------------------
//
//  counter.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package perf

// Stdlib imports.
import (
	"fmt"
	"sync"
	"time"
)


// Counter represents a simple performance counter object. Counters tracking
// an ongoing value while also tracking the min, max and per-second delta
// between samples. Addtionally, statistics can be enabled on a counter object
// to enable tracking of variance, mean, median and standard deviation.
type Counter struct {
	mutex      sync.Mutex
	perSec     float64
	stats      *Stat
	timestamps [2]time.Time
	vals       [2]int64
}

// NewCounter initializes a new Counter object and returns a pointer to it
// for use.
func NewCounter() *Counter {
	newCounter := new(Counter)
	newCounter.Reset()

	return newCounter
}

// Add adds calculates a new counter value by adding the supplied amount to
// the counter's current value.
func (this *Counter) Add(amount int64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	newVal := this.vals[0] + amount

	this.vals[1]       = this.vals[0]
	this.vals[0]       = newVal
	this.timestamps[1] = this.timestamps[0]
	this.timestamps[0] = time.Now()

	// update stats
	if this.stats != nil {
		this.stats.Next(newVal)
	}
}

// DisableStats removes statistical tracking on this counter object.
func (this *Counter) DisableStats() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stats = nil
}

// EnableStats enables statistical tracking on this counter object. Note that stats
// tracking is expensive and should be enabled judiciously on applications that have
// many counters.
func (this *Counter) EnableStats() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stats = NewStat()
}

// Increment calculates and inserts a new NextVal() which is 1 greater
// than the last value.
func (this *Counter) Increment() {
	this.Add(1)
}

// PerSec calculates and returns the per-second derivative from the most recent
// two samples.
func (this *Counter) PerSec() float64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.calcPerSec()
}

// Reset re-initializes the counter object and underlying stats. It does not disable
// stats if they have been enabled on the counter object, only resets them.
func (this *Counter) Reset() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.perSec        = 0
	this.timestamps[1] = time.Now()
	this.timestamps[0] = this.timestamps[1]
	this.vals[1]       = 0
	this.vals[0]       = 0

	if this.stats != nil {
		this.stats.Reset()
	}
}

// Set sets the counter to the supplied value.
func (this *Counter) Set(val int64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.vals[1]       = this.vals[0]
	this.vals[0]       = val
	this.timestamps[1] = this.timestamps[0]
	this.timestamps[0] = time.Now()

	// update stats
	if this.stats != nil {
		this.stats.Next(val)
	}
}

// Stats returns this counter objects statistics object. If stats are
// not enabled, returns nil.
func (this *Counter) Stats() *Stat {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.stats
}

// String pretty-prints the contents of the counter object
func (this *Counter) String() string {
	statTxt := ""

	this.mutex.Lock()
 	if this.stats != nil {
 		statTxt = " " + this.stats.String()
 	}
 	this.mutex.Unlock()

	return fmt.Sprintf(
		"value: %d, perSec: %.2f%s",
		this.Value(),
		this.PerSec(),
		statTxt,
	)
}

// Value returns the current value of the counter.
func (this *Counter) Value() int64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.vals[0]
}

// calcPerSec calculates the rate of change, per-second, since the last value
// arrived.
func (this *Counter) calcPerSec() float64 {
	return float64(this.vals[0] - this.vals[1]) / 
		(this.timestamps[0].Sub(this.timestamps[1])).Seconds()
}

