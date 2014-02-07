//  ---------------------------------------------------------------------------
//
//  web.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package diag

// External imports.
import (
	"github.com/xaevman/goat/lib/perf"
)

// Stdlib imports.
import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

// Diag pages.
var diagUris = []*UriInfo {
	&UriInfo { path: "/diag/blocked", link: "blocked", handler: uriBlocked },
	&UriInfo { path: "/diag/env",     link: "env",     handler: uriEnv     },
	&UriInfo { path: "/diag/mem",     link: "mem",     handler: uriMem     },
	&UriInfo { path: "/diag/perf",    link: "perf",    handler: uriPerf    },
	&UriInfo { path: "/diag/stack",   link: "stack",   handler: uriStack   },
	&UriInfo { path: "/diag/sys",     link: "sys",     handler: uriSys     },
}


// UriInfo represents the data associated with a given Uri in the web server.
type UriInfo struct {
	path    string
	link    string
	handler func(http.ResponseWriter, *http.Request) 
}


// InitWebDiag initializes the web diag uris within an active web server.
func InitWebDiag() {
	runtime.SetBlockProfileRate(1)

	http.HandleFunc("/diag", uriRoot)

	for i := range diagUris {
		uri := diagUris[i]
		http.HandleFunc(uri.path, uri.handler)
	}
}


// uriBlocked is the handler for the /diag/blocked uri.
func uriBlocked(w http.ResponseWriter, req *http.Request) {
	data := NewBlockedData()
	fmt.Fprint(w, data)
}

// uriEnv is the handler for the /diag/env uri.
func uriEnv(w http.ResponseWriter, req *http.Request) {
	data := NewEnvData()
	fmt.Fprintf(w, "%v", data)
}

// uriMem is the handler for the /diag/mem uri.
func uriMem(w http.ResponseWriter, req *http.Request) {
	data := NewMemData()
	fmt.Fprint(w, FmtMemStatsStr(data))
}

// uriPerf is the handler for the /diag/perf uri.
func uriPerf(w http.ResponseWriter, req *http.Request) {
	data := perf.TakeSnapshot()
	fmt.Fprint(w, data.StringBrief())
}

// uriRoot is the handler for the base /diag uri.
func uriRoot(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<div class='header' />")

	fmt.Fprintln(w, "<h1>Debug</h1><hr>")
	fmt.Fprintf(w, "<div class='diagLinks'>")
	for i := range diagUris {
		uri := diagUris[i]
		fmt.Fprintf(w, "<a href=%s>%s</a><br>", uri.path, uri.link)
	}
	fmt.Fprintf(w, "</div>")

	fmt.Fprintln(w, "<br><h1>Profiler</h1><hr>")
	fmt.Fprint(w, "<a href=/debug/pprof/>pprof</a><br>")

	fmt.Fprintf(w, "<div class='footer' />")
}

// uriStack is the handler for the /diag/stack uri.
func uriStack(w http.ResponseWriter, req *http.Request) {
	data := NewFullStackTrace()
	fmt.Fprint(w, data)
}

// uriSys is the handler for the /diag/sys uri.
func uriSys(w http.ResponseWriter, req *http.Request) {
	data := NewSysData()
	fmt.Fprint(w, data.String())
}
