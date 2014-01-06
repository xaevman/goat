//  ---------------------------------------------------------------------------
//
//  config.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package config presents a unified, hierarchical interface for retreiving
// configuration options from any registered providers. The package includes
// builtin providers for retrieving config settings from the environment, 
// ini formatted config files, and xml formatted config files.
package config

// External imports.
import (
	"github.com/xaevman/goat/lib/strutil"
	"github.com/xaevman/goat/core/log"
)

// Stdlib imports.
import(
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"sync"
)

// Common error message format.
const ERR_KEY_NOT_FOUND = "Key not found: %v, default: %v"

// Syncronization helpers.
var mutex sync.Mutex

// Handling of default values.
var (
	m_defaultEntry  = ConfigEntry {
		key: 	"DefaultEntry",
		parser: &m_defaultProvider,
		vals:   []string {},
	}
	m_defaultProvider = defaultProvider{}
)

// Map and priority list of registered clients.
var (
	providerMap = map[string]*list.Element {}
	priList   = list.New()
)

// ConfigProvider defines the interface that should be implemnted by
// config providers.
type ConfigProvider interface {
	GetEntriesByKey(name string) []*ConfigEntry
	GetFirstEntryByKey(name string) *ConfigEntry
	Name() string
	Priority() int
	Shutdown()
}

// GetAllVals searches through all registered config providers and
// returns all values from the first config entry found. If no matching
// config entries are found, it returns the supplied default value.
func GetAllVals(key, defaultVal string) ([]string, *ConfigEntry) {
	entries, err := searchParsers(key)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return []string { defaultVal }, &m_defaultEntry
	}

	vals := entries[0].GetAllVals()
	return vals, entries[0]
}

// GetBoolVal searches through all registered config providers and
// returns a bool value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a bool,
// the supplied default value is returned.
func GetBoolVal(key string, offset int, defaultVal bool) (bool, *ConfigEntry) {
	entries, err := searchParsers(key)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, &m_defaultEntry
	}

	vals         := entries[0].GetAllVals()
	castVal, err := strconv.ParseBool(vals[offset])
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entries[0]
	}

	return castVal, entries[0]
}

// GetEntries returns all matching entries from the first config provider
// with entries that match the supplied key.
func GetEntries(key string) []*ConfigEntry {
	entries, err := searchParsers(key)
	if err != nil {
		return nil
	}

	return entries
}

// GetFloat32Val searches through all registered config providers and
// returns a float32 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a float32,
// the supplied default value is returned.
func GetFloat32Val(key string, offset int, defaultVal float32) (float32, *ConfigEntry) {
	val, entry, err := getFloatVal(key, offset, 32)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return float32(val), entry
}

// GetFloat64Val searches through all registered config providers and
// returns a float64 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a float64,
// the supplied default value is returned.
func GetFloat64Val(key string, offset int, defaultVal float64) (float64, *ConfigEntry) {
	val, entry, err := getFloatVal(key, offset, 64)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return val, entry
}

// GetIntVal searches through all registered config providers and
// returns an int value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to an int,
// the supplied default value is returned.
func GetIntVal(key string, offset, defaultVal int) (int, *ConfigEntry) {
	val, entry, err := getIntVal(key, offset, 0)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return int(val), entry
}

// GetInt8Val searches through all registered config providers and
// returns an int8 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to an int8,
// the supplied default value is returned.
func GetInt8Val(key string, offset int, defaultVal int8) (int8, *ConfigEntry) {
	val, entry, err := getIntVal(key, offset, 8)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return int8(val), entry
}

// GetInt16Val searches through all registered config providers and
// returns an int16 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to an int16,
// the supplied default value is returned.
func GetInt16Val(key string, offset int, defaultVal int16) (int16, *ConfigEntry) {
	val, entry, err := getIntVal(key, offset, 16)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return int16(val), entry
}

// GetInt32Val searches through all registered config providers and
// returns an int32 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to an int32,
// the supplied default value is returned.
func GetInt32Val(key string, offset int, defaultVal int32) (int32, *ConfigEntry) {
	val, entry, err := getIntVal(key, offset, 32)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return int32(val), entry
}

// GetInt64Val searches through all registered config providers and
// returns an int64 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to an int64,
// the supplied default value is returned.
func GetInt64Val(key string, offset int, defaultVal int64) (int64, *ConfigEntry) {
	val, entry, err := getIntVal(key, offset, 64)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return val, entry
}


// GetUintVal searches through all registered config providers and
// returns a uint value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a uint,
// the supplied default value is returned.
func GetUintVal(key string, offset int, defaultVal uint) (uint, *ConfigEntry) {
	val, entry, err := getUintVal(key, offset, 0)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return uint(val), entry
}

// GetUint8Val searches through all registered config providers and
// returns a uint8 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a uint8,
// the supplied default value is returned.
func GetUint8Val(key string, offset int, defaultVal uint8) (uint8, *ConfigEntry) {
	val, entry, err := getUintVal(key, offset, 8)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return uint8(val), entry
}

// GetUint16Val searches through all registered config providers and
// returns a uint16 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a uint16,
// the supplied default value is returned.
func GetUint16Val(key string, offset int, defaultVal uint16) (uint16, *ConfigEntry) {
	val, entry, err := getUintVal(key, offset, 16)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return uint16(val), entry
}

// GetUint32Val searches through all registered config providers and
// returns a uint32 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a uint32,
// the supplied default value is returned.
func GetUint32Val(key string, offset int, defaultVal uint32) (uint32, *ConfigEntry) {
	val, entry, err := getUintVal(key, offset, 32)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return uint32(val), entry
}

// GetUint64Val searches through all registered config providers and
// returns a uint64 value from the given offset of the first valid entry found. If
// no matching entries are found, or the value cannot be correctly parsed to a uint64,
// the supplied default value is returned.
func GetUint64Val(key string, offset int, defaultVal uint64) (uint64, *ConfigEntry) {
	val, entry, err := getUintVal(key, offset, 64)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, entry
	}

	return val, entry
}

// GetVal searches through all registered config providers and
// returns a string value from the given offset of the first valid entry found. If
// no matching entries are found, the supplied default value is returned.
func GetVal(key string, offset int, defaultVal string) (string, *ConfigEntry) {
	entries, err := searchParsers(key)
	if err != nil {
		log.Debug(ERR_KEY_NOT_FOUND, key, defaultVal)
		return defaultVal, &m_defaultEntry
	}

	vals := entries[0].GetAllVals()
	return vals[offset], entries[0]
}

// RegisterConfigProvider registers a new config provider with the config service, 
// also inserting it into its appropriate place in the priority map. Providers
// registered with the same priority level answer in reverse order from which
// they were added.
func RegisterConfigProvider(provider ConfigProvider) {
	mutex.Lock()
	defer mutex.Unlock()

	if priList.Len() < 1 {
		e := priList.PushFront(provider)
		providerMap[provider.Name()] = e
	} else {
		for i := priList.Front(); i != nil; i = i.Next() {
			if i.Next() == nil  {
				e := priList.InsertAfter(provider, i)
				providerMap[provider.Name()] = e
				break
			}

			next := i.Next().Value.(ConfigProvider)
			if next.Priority() >= provider.Priority() {
				e := priList.InsertAfter(provider, i)
				providerMap[provider.Name()] = e
				break
			}
		}
	}

	log.Info("ConfigProvider %v registered", provider.Name())
}

// Shutdown clears all provider registrations and calls Shutdown() on each of them
// in turn. Shutdown order cannot be guaranteed.
func Shutdown() {
	clearRegistrations()
}

// UnregisterConfigProvider removes a registered config provider from the provider
// map and priority list. It will no longer answer GetVal calls from the master config
// service, but can still be used directly unless manually shutdown.
func UnregisterConfigProvider(provider ConfigProvider) {
	mutex.Lock()
	defer mutex.Unlock()

	e := providerMap[provider.Name()]

	priList.Remove(e)
	delete(providerMap, provider.Name())
	
	log.Info("ConfigProvider %v unregistered", provider.Name())
}

// clearRegistrations removes all registered config providers from the provider map
// and priority list, calling Shutdown() on each provider along the way.
func clearRegistrations() {
	mutex.Lock()
	defer mutex.Unlock()

	for _, v := range providerMap {
		parser := v.Value.(ConfigProvider)
		parser.Shutdown()
	}

	for _, v := range providerMap {
		parser := v.Value.(ConfigProvider)
		delete(providerMap, parser.Name())
	}

	priList.Init()
}

// getFloatVal is the internal function backing the public GetFloatVal functions. It
// searches all registered parsers for relevant entries and returns the first one found.
func getFloatVal(key string, offset int, bitsize int) (float64, *ConfigEntry, error) {
	entries, err := searchParsers(key)
	if err != nil {
		return 0.0, &m_defaultEntry, err
	}

	vals         := entries[0].GetAllVals()
	castVal, err := strconv.ParseFloat(vals[offset], bitsize)
	if err != nil {
		return 0.0, entries[0], err
	}

	return castVal, entries[0], nil
}

// getIntVal is the internal function backing the public GetIntVal functions. It
// searches all registered parsers for relevant entries and returns the first one found.
func getIntVal(key string, offset, bitsize int) (int64, *ConfigEntry, error) {
	entries, err := searchParsers(key)
	if err != nil {
		return 0, &m_defaultEntry, err
	}

	vals         := entries[0].GetAllVals()
	castVal, err := strconv.ParseInt(vals[offset], 10, bitsize)
	if err != nil {
		return 0, entries[0], err
	}

	return castVal, entries[0], nil
}

// getUintVal is the internal function backing the public GetUintVal functions. It
// searches all registered parsers for relevant entries and returns the first one found.
func getUintVal(key string, offset, bitsize int) (uint64, *ConfigEntry, error) {
	entries, err := searchParsers(key)
	if err != nil {
		return 0, &m_defaultEntry, err
	}

	vals         := entries[0].GetAllVals()
	castVal, err := strconv.ParseUint(vals[offset], 10, bitsize)
	if err != nil {
		return 0, entries[0], err
	}

	return castVal, entries[0], nil
}

// searchParsers loops through registered config providers in priority order, looking
// for matching entries, and returns all entries from the first config provider
// that can fulfill the request.
func searchParsers(key string) ([]*ConfigEntry, error) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := priList.Front(); i != nil; i = i.Next() {
		parser  := i.Value.(ConfigProvider)
		log.Debug("Searching parser %v for %v", parser.Name(), key)
		
		entries := parser.GetEntriesByKey(key)
		if entries != nil {
			return entries, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Entry not found (%v)", key))
}

// ConfigEntry objects represent a given key and its associated values.
type ConfigEntry struct {
	key    string
	parser ConfigProvider
	vals   []string
}

// GetAllVals returns all values associated with this config entry.
func (this *ConfigEntry) GetAllVals() []string {
	return this.vals
}

// GetVal attempts to return the value at a given offset with the list of this
// entries values. If offset is out of range, returns an empty string.
func (this *ConfigEntry) GetVal(offset int) string {
	if offset > len(this.vals) - 1 {
		return ""
	}

	return this.vals[offset]
}

// Len returns the number of values associated with this config entry.
func (this *ConfigEntry) Len() int {
	return len(this.vals)
}

// Name returns the name, or key, of this config entry.
func (this *ConfigEntry) Name() string {
	return this.key
}

// Parser returns the parent ConfigProvider instance that this config entry is a child
// of.
func (this *ConfigEntry) Parser() ConfigProvider {
	return this.parser
}

// String returns a nicely formatted string representing the ConfigEntry object.
func (this *ConfigEntry) String() string {
	return fmt.Sprintf(
		"%v::%v = %v", 
		this.parser.Name(),
		this.key,
		strutil.StrArrayToCsv(this.vals),
	)
}

// defaultProvider is returned as the ConfigProvider of record when the config system
// has to fall back to given default vals for a query.
type defaultProvider struct {}

func (this *defaultProvider) GetEntriesByKey(name string) []*ConfigEntry { 
	return nil
}
func (this *defaultProvider) GetFirstEntryByKey(name string) *ConfigEntry {
	return nil
}
func (this *defaultProvider) Name() string  { return "DefaultProvider"}
func (this *defaultProvider) Priority() int { return -1 }
func (this *defaultProvider) Shutdown() {}
