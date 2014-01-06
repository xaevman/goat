//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package config

import(
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	InitEnvProvider(1)
	InitIniProvider("./test.ini", 2)

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

func printConfig(key string, vals []string, parser ConfigProvider) {
	if vals == nil {
		return
	}

	for i, v := range vals {
		log.Printf("%v.%v[%v]: %v", parser.Name(), key, i, v)
	}
}
