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

// Stdlib imports.
import (
    "bytes"
    "sort"
    "sync"
)

// CounterSet registry and synchronization objects
var (
    mutex   sync.RWMutex
    perfMap = make(map[string]*CounterSet, 0)
)

// DumpString dumps every value, of every Counter, of every CounterSet registered
// with the perf service, in string format.
func DumpString() string {
    var buffer bytes.Buffer

    counterSets := GetAllCounterSets()

    for i := 0; i < len(counterSets); i++ {
        buffer.WriteString(counterSets[i].String())
    }

    return buffer.String()
}

// GetAllCounterSets returns a slice with pointers to all registered CounterSet
// objects.
func GetAllCounterSets() []*CounterSet {
    mutex.RLock()
    defer mutex.RUnlock()

    cursor      := 0
    counterList := make([]*CounterSet, len(perfMap))
    sKeys       := make([]string, len(perfMap))

    for k, _ := range perfMap {
        sKeys[cursor] = k
        cursor++
    }

    sort.Strings(sKeys)

    for i := range sKeys {
        counterList[i] = perfMap[sKeys[i]]
    }

    return counterList
}

// GetCounterSet returns the named CounterSet object, if one is registered
// by that name. Otherwise, nil is returned.
func GetCounterSet(name string) *CounterSet {
    mutex.RLock()
    defer mutex.RUnlock()

    return perfMap[name]
}

// registerCounterSet adds a new CounterSet object to the registry, overwriting
// any previous objects that were registered with that name.
func registerCounterSet(perfs *CounterSet) {
    mutex.Lock()
    defer mutex.Unlock()

    perfMap[perfs.name] = perfs
}

// unregisterCounterSet removes the named CounterSet object from the registry.
func unregisterCounterSet(name string) {
    mutex.Lock()
    defer mutex.Unlock()

    delete(perfMap, name)
}
