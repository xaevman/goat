//  ---------------------------------------------------------------------------
//
//  gobuild.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// gobuild recursively searches a path for go packages with a 'BUILD'
// file in their directory and builds the package or application. Builds
// are performed for each platform, and with the options specified in 
// the xml of that package's BUILD file.
package main

// External imports
import (
	"github.com/xaevman/goat/lib/fsutil"
)

// Stdlib imports
import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

const buildFileName = "BUILD"

type build struct {
	Dir string
	Path string
	Platforms []platform `xml:"Platform"`
}

type platform struct {
	OS string `xml:"os,attr"`
	Arch string `xml:"arch,attr"`
}

func main() {
	// startup
    fmt.Println("Start run...")
    start := time.Now()

    // kick it off
    startDir := "./"
    if len(os.Args) > 1 {
    	startDir = os.Args[1]
    }

    rootDir := fsutil.NewSearchDir(10, 10)

    go rootDir.SearchFiles(startDir, buildFileName)

    buildList := make([]string, 0)
    run       := true

    // wait for search results
    for run {
	    select {
	    case a := <-rootDir.FileChan:
    		fmt.Printf("BUILD detected (%v)\n", a)
    		buildList = append(buildList, a)
	    case a := <-rootDir.ErrChan:
	    	fmt.Println(a)
	    case <-rootDir.DoneChan:
	    	run = false
	    }
	}

	// perform all the builds
	buildChan := make(chan bool)
	for _, item := range buildList {
    	buildData, err := ioutil.ReadFile(item)
		if err != nil {
			fmt.Printf("Error opening BUILD file %v (%v)\n", item, err)
			continue
		}
		
		b := build { 
			Dir: path.Dir(item), 
			Path: item, 
		}
		
		xml.Unmarshal(buildData, &b)
		go goBuild(b, buildChan)
	}

	for i := 0; i < len(buildList); i++ {
		<-buildChan
	}

	// print results
    fmt.Println("Run complete")
    fmt.Printf(
    	"%v builds in %v\n",
    	len(buildList),
		time.Since(start),
	)
}

func goBuild(b build, c chan bool) {
	pChan := make(chan bool)
	for _, platform := range b.Platforms {
		go goBuildPlatform(b, platform, pChan)
	}

	for i := 0; i < len(b.Platforms); i++ {
		<-pChan
	}

	c<- true
}

func goBuildPlatform(b build, p platform, c chan bool) {
	os.Setenv("GOOS",   p.OS)
	os.Setenv("GOARCH", p.Arch)
	os.Chdir(b.Dir)

	fmt.Printf(
		"Starting build %v (%v-%v)\n", 
		b.Dir,
		p.OS,
		p.Arch,
	)

	cmd         := exec.Command("go", "install")
	start       := time.Now()
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running %v (%v)\n", cmd.Path, err)
		fmt.Printf("\t%v\n", string(output))
	}

	fmt.Printf(
		"%v (%v-%v) complete\t%v\n", 
		b.Dir, 
		p.OS, 
		p.Arch, 
		time.Since(start),
	)

	c<- true
}
