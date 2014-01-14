//  ---------------------------------------------------------------------------
//
//  perfs.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

import (
	"bytes"
	"fmt"
	"sync/atomic"
)

type PerfCounters struct {
	nameMap map[int]string
	perfs   []uint64
}

func NewPerfs(size int, names []string) *PerfCounters {
	if len(names) != size {
		return nil
	}

	p := PerfCounters {
		nameMap: make(map[int]string, size),
		perfs:   make([]uint64, size),
	}

	for i := range names {
		p.nameMap[i] = names[i]
	}

	return &p
}

func (this *PerfCounters) Get(offset int) uint64 {
	if offset > len(this.perfs) {
		return 0
	}

	return atomic.LoadUint64(&this.perfs[offset])
}

func (this *PerfCounters) Increment(offset int) {
	if offset > len(this.perfs) {
		return
	}

	atomic.AddUint64(&this.perfs[offset], 1)
}

func (this *PerfCounters) Reset() {
	for i := range this.perfs {
		atomic.StoreUint64(&this.perfs[i], 0)
	}
}

func (this *PerfCounters) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("Perf dump")

	for i := range this.perfs {
		buffer.WriteString(fmt.Sprintf(
			"\n%v = %v",
			this.nameMap[i],
			this.perfs[i],
		))
	}

	return buffer.String()
}
