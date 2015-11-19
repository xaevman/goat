//  ---------------------------------------------------------------------------
//
//  main.go
//
//  Copyright (c) 2014, Jared Chavez.
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
//  Usage: gobuild <search root directory>
package main

// External imports
import (
    "github.com/xaevman/goat/lib/fs"
)

// Stdlib imports
import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path"
    "path/filepath"
    "runtime"
    "strings"
    "time"
)

// File name to search for in the target path.
const buildFileName = "BUILD"

// Flag which specifies whether or not to install the build result for the
// executing platform and architecture.
var install = false

// build represents a build node in the BUILD xml file.
type build struct {
    Dir       string
    Install   bool       `xml:"install,attr"`
    Name      string
    Path      string
    Platforms []platform `xml:"Platform"`
}

// platform represents a platform node in the BUILD xml file.
type platform struct {
    OS   string `xml:"os,attr"`
    Arch string `xml:"arch,attr"`
}

// main is the entry point for the application. The first argument passed
// on the command line is used as the base path to start the recursive
// search.
func main() {
    // startup
    fmt.Println("Start run...")
    start := time.Now()

    // kick it off
    startDir := "./"
    if len(os.Args) > 1 {
        startDir = os.Args[1]
    }

    rootDir := fs.NewSearchDir()

    go rootDir.SearchFiles(startDir, buildFileName)

    buildList := make([]string, 0)

    // wait for search results
    func() {
        for {
            select {
            case a := <-rootDir.FileChan:
                fmt.Printf("BUILD detected (%v)\n", a)
                buildList = append(buildList, a)
            case a := <-rootDir.ErrChan:
                fmt.Println(a)
            case <-rootDir.DoneChan:
                return
            }
        }
    }()

    // perform all the builds
    buildChan := make(chan bool)
    for _, item := range buildList {
        buildData, err := ioutil.ReadFile(item)
        if err != nil {
            fmt.Printf("Error opening BUILD file %v (%v)\n", item, err)
            continue
        }

        b := &build {
            Dir:     path.Dir(item),
            Install: false,
            Path:    item,
        }

        resolveName(b)

        xml.Unmarshal(buildData, b)
        goBuild(b, buildChan)
    }

    // print results
    fmt.Println("Run complete")
    fmt.Printf(
        "%v builds in %v\n",
        len(buildList),
        time.Since(start),
    )
}

// resolveName attempts to resolve the name of project directory. This name
// is later used as the file name of the resulting binary.
func resolveName(b *build) {
    absPath, err := filepath.Abs(b.Dir)
    if err != nil {
        fmt.Printf("Can't resolve name for directory %s\n", b.Dir)
        return
    }

    fileName := filepath.Base(absPath)
    fmt.Printf("Resolved fileName = %s\n", fileName)
    b.Name = fileName
}

// goBuild takes a given build and stars a separate go routine to build
// each of the targetted platforms.
func goBuild(b *build, c chan bool) {
    pChan := make(chan bool)
    for _, platform := range b.Platforms {
        go goBuildPlatform(b, platform, pChan)
    }

    for i := 0; i < len(b.Platforms); i++ {
        <-pChan
    }
}

// goBuildPlatform takes a given set of build settings and target platform
// and executes to the build with those settings.
func goBuildPlatform(b *build, p platform, c chan bool) {
    fmt.Printf(
        "Starting build %v (%v-%v)\n",
        b.Dir,
        p.OS,
        p.Arch,
    )

    // build paths
    extension := ""
    if p.OS == "windows" {
        extension = ".exe"
    }

    gopath   := fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH"))
    goos     := fmt.Sprintf("GOOS=%s", p.OS)
    goarch   := fmt.Sprintf("GOARCH=%s", p.Arch)
    destPath := fmt.Sprintf("bin/%s/%s", p.OS, p.Arch)
    destFile := fmt.Sprintf("%s/%s%s", destPath, b.Name, extension)

    // build
    cmd         := exec.Command("go", "build", "-o", destFile)
    cmd.Dir      = b.Dir
    cmd.Env      = []string { gopath, goos, goarch }
    start       := time.Now()
    output, err := cmd.CombinedOutput()

    if err != nil {
        fmt.Printf("Error running %v (%v)\n", cmd.Path, err)
        fmt.Printf("\t%v\n", string(output))
    }

    // install if requested
    if b.Install               == true &&
       strings.ToLower(p.OS)   == runtime.GOOS &&
       strings.ToLower(p.Arch) == runtime.GOARCH {
        fmt.Printf("Installing %s-%s...\n", p.OS, p.Arch)
        cmd         = exec.Command("go", "install")
        cmd.Dir     = b.Dir
        start       = time.Now()
        output, err = cmd.CombinedOutput()

        if err != nil {
            fmt.Printf("Error installing %v (%v)\n", cmd.Path, err)
            fmt.Printf("\t%v\n", string(output))
        }
    }

    // report completion
    fmt.Printf(
        "%v (%v-%v) complete\t%v\n",
        b.Dir,
        p.OS,
        p.Arch,
        time.Since(start),
    )

    c <- true
}
