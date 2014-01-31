//  ---------------------------------------------------------------------------
//
//  iniprovider.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package config

// External imports.
import (
	"github.com/xaevman/goat/lib/fs"
	"github.com/xaevman/goat/lib/perf"
	"github.com/xaevman/goat/lib/str"
	"github.com/xaevman/goat/core/log"
)

// Stdlib imports.
import(
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Perf counters.
const (
	PERF_CFG_INI_PRIORITY = iota
	PERF_CFG_INI_QUERIES
	PERF_CFG_INI_COUNT
)

// Perf counter friendly names.
var cfgIniPerfNames = []string {
	"Priority",
	"Queries",
}

// The base directory from which all supplied file paths will be built.
// Defaults to the directory of the primary executable.
var IniDir = filepath.Dir(fs.ExeFile())

// Ini provider module name
const INI_MOD_NAME = "IniProvider"

// InitIniProvider initializes a new IniProvider config provider for the
// give path, registers it with the conig service, and returns a 
// pointer to the object for direct use, if required.
func InitIniProvider(path string, pri int) *IniProvider {
	fullPath     := filepath.Join(IniDir, path)
	exists, info := fs.FileExists(fullPath)

	if !exists{
		log.Error("Config file doesn't exist %v", fullPath)
		return nil
	}

	if info.IsDir() {
		log.Error("Config path points to a directory (%v)", fullPath)
		return nil
	}

	fileName    := filepath.Base(fullPath)
	name        := fmt.Sprintf("%v.%v", INI_MOD_NAME, fileName)
	iniProvider := IniProvider {
		entries    : make(map[string][]*ConfigEntry, 0),
		filePath   : fullPath,
		moduleName : name,
		perfs      : perf.NewCounterSet(
			"Module.Config." + name,
			PERF_CFG_INI_COUNT,
			cfgIniPerfNames,
		),
		priority   : pri,
	}

	err := iniProvider.parseConfig()
	if err != nil {
		log.Error("Unable to open ini file %v (%v)", fullPath, err)
		return nil
	}

	iniProvider.perfs.Set(PERF_CFG_INI_PRIORITY, int64(pri))

	RegisterConfigProvider(&iniProvider)

	return &iniProvider
}

// IniProvider represents a ConfigProvider implementation which can query
// a given ini-formatted config file for config entries.
type IniProvider struct {
	entries    map[string][]*ConfigEntry
	filePath   string
	moduleName string
	perfs      *perf.CounterSet
	priority   int
}

// GetEntriesByKey returns all entries within the ini file which match
// the queried key name. ConfigEntry names follow the format
// <Section>.<Key> .
func (this *IniProvider) GetEntriesByKey(name string) []*ConfigEntry {
    list    := this.entries[name]
    if list == nil {
    	return nil
    }

	results := make([]*ConfigEntry, 0)

	for _, v := range list {
		results = append(results, v)
	}

	this.perfs.Increment(PERF_CFG_INI_QUERIES)

	return results
}

// GetFirstEntryByKey returns the first entry within the ini file which matches
// the queried key name. ConfigEntry names follow the format <Section>.<Key> .
func (this *IniProvider) GetFirstEntryByKey(name string) *ConfigEntry {
	entries := this.GetEntriesByKey(name)
	if entries == nil || len(entries) < 1 {
		return nil
	}

	this.perfs.Increment(PERF_CFG_INI_QUERIES)

	return entries[0]
}

// Name returns "IniProvider", the name of this config module.
func (this *IniProvider) Name() string {
	return this.moduleName
}

// Priority returns the assigned priority for this EnvParser object.
func (this *IniProvider) Priority() int {
	return this.priority
}

// Unused.
func (this *IniProvider) Shutdown() {}

// IsIniCommentLine tests whether or not a given string is an ini-style
// comment line.
func isCommentLine(line string) bool {
	if len(line) < 1 {
		return true
	}

	if strings.Index(line, ";") == 0 {
		return true
	}

	return false
}

// IsIniSectionLine tests whether or not a given string is a line denoting
// an ini-style section.
// 	[Example.Section]
func isSectionLine(line string) (bool, string) {
	exp := regexp.MustCompile("^\\s*\\[(.*)\\]\\s*$")
	result := exp.FindStringSubmatch(line)

	if result == nil {
		return false, line
	}

	return true, strings.TrimSpace(result[1])
}

// newEntry takes a section name and an ini-style config line, formats
// it as a ConfigEntry object, and adds it to the list of ConfigEntries
// associated with this parser.
func (this *IniProvider) newEntry(section, line string) {
	line = trimCommentText(line)

	pair := strings.Split(line, "=")
	if len(pair) != 2 {
		return
	}

	keyName := fmt.Sprintf("%v.%v", section, strings.TrimSpace(pair[0]))
	valList := str.DelimToStrArray(pair[1], ",")

	cfgEntry := ConfigEntry {
		key: keyName,
		parser: this,
		vals: valList,
	}

	this.entries[cfgEntry.key] = append(this.entries[cfgEntry.key], &cfgEntry)
	log.Debug(
		"%v:%v:%v", 
		len(this.entries),
		len(this.entries[cfgEntry.key]), 
		&cfgEntry,
	)
}

// parseConfig opens a given ini config file and Marshals it into an IniProvider
// object representation.
func (this *IniProvider) parseConfig() error {
	file, err := fs.OpenFile(this.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var section string

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break	// eof
		}

		line = strings.TrimSpace(line)

		if isCommentLine(line) {
			continue
		}

		isSection, line := isSectionLine(line)
		if isSection {
			section = line
			continue
		}

		this.newEntry(section, line)
	}

	return nil
}

// TrimIniCommentText trims any comment text out of a given string.
// 	key1 = val1 ;this text will be trimmed in resulting string
func trimCommentText(line string) string {
	i := strings.Index(line, ";")
	if i < 0 {
		return line
	}

	return line[0:i]
}
