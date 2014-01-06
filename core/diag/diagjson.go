//  ---------------------------------------------------------------------------
//
//  diagjson.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package diag

import(
	"encoding/json"
)

// FmtDiagStr aggregates and returns diagnostics information in json format.
// Diagnostic information includes hostname, CPU count, environment data,
// stack traces for all running goroutines, and memory allocation statistics.
func FmtDiagJson(err error) string {
	data    := New(err)
	json, _ := json.MarshalIndent(data, "", "    ")
	
	return string(json)
}
