//  ---------------------------------------------------------------------------
//
//  envprovider.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package config

// External imports
import (
    "github.com/xaevman/goat/lib/perf"
    "github.com/xaevman/goat/lib/str"
)

// Stdlib imports
import(
    "os"
    "sync"
)

// Perf counters.
const (
    PERF_CFG_ENV_PRIORITY = iota
    PERF_CFG_ENV_QUERIES
    PERF_CFG_ENV_COUNT
)

// Perf counter friendly names.
var cfgEnvPerfNames = []string {
    "Priority",
    "Queries",
}

// Environment provider module name.
const ENV_MOD_NAME = "EnvProvider"

// Static instance and synchronization
var envMutex    sync.Mutex
var envProvider *EnvProvider

// InitEnvProvider initializes a new EnvProvider config provider, registers it with
// the config services, and returns a pointer to the object for direct use,
// if required.
func InitEnvProvider(pri int) *EnvProvider {
    envMutex.Lock()
    defer envMutex.Unlock()

    if envProvider == nil {
        envProvider = &EnvProvider {
            moduleName : ENV_MOD_NAME,
            perfs      : perf.NewCounterSet(
                "Module.Config." + ENV_MOD_NAME,
                PERF_CFG_ENV_COUNT,
                cfgEnvPerfNames,
            ),
            priority   : pri,
        }
    }

    if pri == envProvider.Priority() {
        return envProvider
    }

    UnregisterConfigProvider(envProvider)
    envProvider.priority = pri

    envProvider.perfs.Set(PERF_CFG_ENV_PRIORITY, int64(pri))

    RegisterConfigProvider(envProvider)

    return envProvider
}

// EnvProvider represents a ConfigProvider implementation that queries the
// system environment for config entries.
type EnvProvider struct {
    moduleName string
    perfs      *perf.CounterSet
    priority   int
}

// GetEntriesByKey returns the requested environment variable, if present, 
// formatted as a ConfigEntry object. If an environment variable matching
// requested name doesn't exist, nil is returned. Only one entry will
// ever be returned from this call, despite its signature, since duplicate
// environment variables cannot exist.
func (this *EnvProvider) GetEntriesByKey(name string) []*ConfigEntry {
    entry := this.GetFirstEntryByKey(name)
    if entry == nil {
        return nil
    }

    this.perfs.Increment(PERF_CFG_ENV_QUERIES)

    return []*ConfigEntry { entry }
}

// GetFirstEntryByKey returns the requested environment variable, if present,
// formatted as a ConfigEntry object. If an environment variable matching
// the requested name doesn't exist, nil is returned.
func (this *EnvProvider) GetFirstEntryByKey(name string) *ConfigEntry {
    val := os.Getenv(name)
    if val == "" {
        return nil
    }

    this.perfs.Increment(PERF_CFG_ENV_QUERIES)

    return newEnvEntry(name, val, this)
}

// Name returns "EnvProvider", the name of this config module.
func (this *EnvProvider) Name() string {
    return this.moduleName
}

// Priority returns the assigned priority for this EnvProvider object.
func (this *EnvProvider) Priority() int {
    return this.priority
}

// Unused in this module.
func (this *EnvProvider) Shutdown() {}

// newEnvEntry creates a ConfigEntry object, populates it with values
// from the system environment, and returns a pointer to the object for use.
func newEnvEntry(name, val string, parent *EnvProvider) *ConfigEntry {
    valList := str.DelimToStrArray(val, string(os.PathListSeparator))
    entry   := ConfigEntry {
        key:    name,
        parser: parent,
        vals:   valList,
    }

    return &entry
}
