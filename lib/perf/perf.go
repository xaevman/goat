//  ---------------------------------------------------------------------------
//
//  perf.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package perf exposes a centralized API for storing, tracking, and querying
// performance data for a Go application.
package perf

import (
	"sync"
)

// Counters registry and synchronization objects
var (
	mutex   sync.RWMutex
	perfMap = make(map[string]*Counters, 0)
)

// GetAllPerfs returns a slice with pointers to all registered Counters
// objects.
func GetAllPerfs() []*Counters {
	mutex.RLock()
	defer mutex.RUnlock()

	cursor      := 0
	counterList := make([]*Counters, len(perfMap))

	for k, _ := range perfMap {
		counterList[cursor] = perfMap[k]
		cursor++
	}

	return counterList
}

// GetPerfs returns the named Counters object, if one is registered
// by that name. Otherwise, nil is returned.
func GetPerfs(name string) *Counters {
	mutex.RLock()
	defer mutex.RUnlock()

	return perfMap[name]
}

// registerPerfs adds a new Counters object to the registry, overwriting
// any previous objects that were registered with that name.
func registerPerfs(perfs *Counters) {
	mutex.Lock()
	defer mutex.Unlock()

	perfMap[perfs.name] = perfs
}

// unregisterPerfs removes the named Counters object from the registry.
func unregisterPerfs(name string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(perfMap, name)
}
