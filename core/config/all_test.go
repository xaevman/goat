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

// Package config presents a unified, hierarchical interface for retreiving
// configuration options from any registered providers. The package includes
// builtin providers for retrieving config settings from the environment and
// ini formatted config files.
package config

// Stdlib imports.
import(
	"log"
	"testing"
)

// TestConfig initializes a couple of ini providers and
// attempts to pull values back out of them.
func TestConfig(t *testing.T) {
	IniDir = "./"

	InitEnvProvider(1)
	if InitIniProvider("test.ini", 2) == nil {
		t.Fatal("Ini file not found")
	}

	key         := "PATH"
	data, entry := GetAllVals(key, "/bin:/sbin")
	printConfig(key, data, entry.Parser())

	key         = "Ini.Section.key1"
	data, entry = GetAllVals(key, "default1")
	printConfig(key, data, entry.Parser())

	key         = "This.Key.Shouldnt.exist"
	data, entry = GetAllVals(key, "default3")
	printConfig(key, data, entry.Parser())
}

//printConfig prints the value data retreived from the config system.
func printConfig(key string, vals []string, parser ConfigProvider) {
	if vals == nil {
		return
	}

	for i, v := range vals {
		log.Printf("%v.%v[%v]: %v", parser.Name(), key, i, v)
	}
}
