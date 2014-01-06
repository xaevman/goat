//  ---------------------------------------------------------------------------
//
//  dir.go
//
//  Written by Jared Chavez (2014-01-05)
//  Owned by Jared Chavez <xaevman@gmail.com>
//
//  Copyright (c) 2014 Jared Chavez
//
//  -----------

package main

import (
	"fmt"
	"github.com/xaevman/goat/lib/fsutil"
	"os"
)

var dirSearch *fsutil.SearchDir

func main() {
	// kick it off
    startDir := "./"
    if len(os.Args) > 1 {
    	startDir = os.Args[1]
    }

    filter := "*"
    if len(os.Args) > 2 {
    	filter = os.Args[2]
    }

	dirSearch = fsutil.NewSearchDir(0, 0)
	go dirSearch.Search(startDir, filter)

	handleResults()
}

func handleResults() {
	loop := true

	// loop until search completes
	for loop {
		select {
		case err := <-dirSearch.ErrChan:
			fmt.Printf("Error: %v\n", err)
		case match := <-dirSearch.FileChan:
			fmt.Println(match)
		case <-dirSearch.DoneChan:
			loop = false;
		}
	}

	// empty all buffers
	for {
		select {
		case err := <-dirSearch.ErrChan:
			fmt.Printf("Error: %v\n", err)
		case match := <-dirSearch.FileChan:
			fmt.Println(match)
		default:
			return
		}
	}
}