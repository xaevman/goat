//  ---------------------------------------------------------------------------
//
//  all_test.go
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
	"log"
	"testing"
	"time"
)

// TestStopwatch instantiates a Stopwatch object and does a few timing tests,
// making sure that the measured times are within a resaonable range of elapsed
// time.
func TestStopwatch(t *testing.T) {
	s := new(Stopwatch)

	s.Start()

	<-time.After(15 * time.Millisecond)
	x := s.MarkMs()
	if x < 14 || x > 16 {
		t.Fatalf("Measured ms after 15ms is out of range (%d)", x)
	}

	log.Println("15ms: passed")

	<-time.After(445 * time.Millisecond)
	x = s.MarkMs()
	if x < 459 || x > 462 {
		t.Fatalf("Measured ms after 460ms is incorrect (%d)", x)
	}

	log.Println("460ms: passed")

	s.Restart()
	<-time.After(1 * time.Second)
	x = s.MarkSec()
	if x != 1 {
		t.Fatalf("Measured sec after 1sec is incorrect(%d)", x)
	}

	log.Println("1sec: passed")

	<-time.After(4 * time.Second)
	x = s.MarkSec()
	if x != 5 {
		t.Fatalf("Measured sec after 5sec is incorrect(%d)", x)
	}

	log.Println("5sec: passed")

	<-time.After(2 * time.Second)
	s.Stop()

	<-time.After(3 * time.Second)
	x = s.MarkSec()
	if x != 7 {
		t.Fatalf("Measured sec after 7sec is incorrect(%d)", x)
	}

	// just make sure this doesn't crash
	s.Mark()

	log.Println("7sec: passed")
}
