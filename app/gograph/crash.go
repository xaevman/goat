//  ---------------------------------------------------------------------------
//
//  crash.go
//
//  Copyright (c) 2014, Jared Chavez.
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package main

// External imports.
import (
    "github.com/xaevman/goat/mod/diag"
    "github.com/xaevman/goat/mod/log"
)

// GoGraphCrash is a goapp.CrashHandler implementation for a 
// GoGraph instance.
type GoGraphCrash struct {}

// OnCrash logs the crash text and also dumps a stack trace to 
// the error log for debugging.
func (this *GoGraphCrash) OnCrash(crashData interface{}) {
    log.Error("%s", crashData)
    log.Error("\n%s", diag.NewStackString())
}
