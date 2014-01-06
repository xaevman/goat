//  ---------------------------------------------------------------------------
//
//  diagjson.go
//
//  Written by Jared Chavez (2014-01-01)
//  Owned by Jared Chavez <xaevman@gmail.com>
//
//  Copyright (c) 2014 Jared Chavez
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
