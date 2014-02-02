//  ---------------------------------------------------------------------------
//
//  dir.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package fs

// Stdlib imports.
import (
	"fmt"
	"os"
	"path/filepath"
)

// DirExists tests to see if a given directory exists on disk and returns
// true, along with a relevant os.FileInfo object if so.
func DirExists(path string) (bool, os.FileInfo) {
	exists, fileInfo := FileExists(path)
	if !exists {
		return false, nil
	}

	if !fileInfo.IsDir() {
		return false, nil
	}

	return true, fileInfo
}

// Mkdir wraps os.Mkdir with a simple DirExists check for more succinct use
// in making sure a directory should exist.
func Mkdir(path string, perm os.FileMode) error {
	exists, _ := DirExists(path)
	if !exists {
		err := os.MkdirAll(path, perm)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewSearchDir is a helper function which returns a pointer to a new SearchDir
// object with the Err and File channels set to the specified buffer sizes.
func NewSearchDir() *SearchDir {
	newObj := SearchDir {
		DoneChan: make(chan bool),
		ErrChan:  make(chan string),
		FileChan: make(chan string),
	}

	return &newObj
}

// SearchDir encapsulates file system search functionality along with 
// 3 communication channels for use in concurrent code. DoneChan is written to 
// when a search operation is complete. ErrChan is written to when any errors 
// are encountered during the search. FileChan is written to for each matching 
// file or directory.
type SearchDir struct {
	DoneChan chan bool
	ErrChan  chan string
	FileChan chan string
}

// Search performs a recursive file system search for files and directories
// starting in the specified root directory, matching the given pattern.
func (this *SearchDir) Search(root, pattern string) {
	this.search(root, pattern, true, true)
}

// SearchDirs performs a recursive file system search for directories
// starting in the specified root directory, matching the given pattern.
func (this *SearchDir) SearchDirs(root, pattern string) {
	this.search(root, pattern, false, true)
}

// SearchFiles performs a recursive file system search for files
// starting in the specified root directory, matching the given pattern.
func (this *SearchDir) SearchFiles(root, pattern string) {
	this.search(root, pattern, true, false)
}


// search does a filepath.Walk starting at the specified root directory,
// sending any events back to the appropriate channels.
func (this *SearchDir) search(root, pattern string, files bool, dirs bool) {
	err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			// err chek
			if err != nil {
				return err
			}

			// type matching
			if info.IsDir() && !dirs {
				return nil
			}

			if !info.IsDir() && !files {
				return nil
			}

			// name matching
			matched, err := filepath.Match(pattern, info.Name())
			if !matched {
				return nil
			}

			if err != nil {
				return err
			}

			// relevant result
			this.FileChan <- path
			return nil
		},
	)

	if err != nil {
		this.ErrChan <- fmt.Sprintf(
			"SearchDir.search error: %v (%v)",
			root,
			err,
		)
	}

	this.DoneChan <- true
}
