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

// Package fs provides helper functions to assist with common operations
// on files and directories.
package fs

// Stdlib imports.
import(
	"log"
	"os"
	"path/filepath"
	"testing"
)

// Available search types.
const (
	All = iota
	DirsOnly
	FilesOnly
)

// Base directory for testing
var searchRoot string = filepath.Join(".", "tests")

// DirectorySearcher to use during search tests.
var dir = SearchDir {
	DoneChan: make(chan bool),
	ErrChan:  make(chan string),
	FileChan: make(chan string),
}

// List of directories to test against.
var dirs = []string {
	filepath.Join(searchRoot, "d1"),
	filepath.Join(searchRoot, "d1", "d1A"),
	filepath.Join(searchRoot, "d1", "d1B"),
	filepath.Join(searchRoot, "d1", "d1B", "d1Ba"),
	filepath.Join(searchRoot, "d1", "d1C"),
	filepath.Join(searchRoot, "d2"),
	filepath.Join(searchRoot, "d2", "d2A", "d2Aa", "d2Aa_"),
	filepath.Join(searchRoot, "d3"),
	filepath.Join(searchRoot, "c3"),
	filepath.Join(searchRoot, "d4"),
	filepath.Join(searchRoot, "d5"),
}

// List of files to test against.
var files = []string {
	filepath.Join(searchRoot, "d1", "file1.f"),
	filepath.Join(searchRoot, "d1", "file2.f"),
	filepath.Join(searchRoot, "d1", "d1B", "j78924uoiejd.f"),
	filepath.Join(searchRoot, "d1", "d1B", "d1Ba", "24fdf.f"),
	filepath.Join(searchRoot, "d1", "d1C", "file1.f"),
	filepath.Join(searchRoot, "d2", "file1.f"),
	filepath.Join(searchRoot, "d2", "d2A", "d2Aa", "d2Aa_", "xf4gtgfd.f"),
	filepath.Join(searchRoot, "d3", "rawr.f"),
	filepath.Join(searchRoot, "d4", "chugalugga.f"),
	filepath.Join(searchRoot, "d5", "filexxx.f"),
	filepath.Join(searchRoot, "d1", "d1B", "file9.f"),
	filepath.Join(searchRoot, "d1", "d1B", "d1Ba", "newfile.f"),
	filepath.Join(searchRoot, "d2", "file2.f"),
	filepath.Join(searchRoot, "d2", "d2A", "d2Aa", "d2Aa_", "file1.f"),
	filepath.Join(searchRoot, "d3", "fsghfd.f"),
	filepath.Join(searchRoot, "d4", "tests.f"),
	filepath.Join(searchRoot, "d5", "file1.f"),
}

// List of bad files to test against.
var badFiles = []string {
	filepath.Join(".", "d1", "noexist", "f45gdff.f"),
	filepath.Join(".", "d1", "d1A", "test123", "newfile.1"),
	filepath.Join(".", "d1", "d1C", "blah", "test.f"),
}

// TestInit initializes the test by removing any remnants from a previous run
// and recreating the ./tests base directory.
func TestInit(t *testing.T) {
	err := os.RemoveAll(searchRoot)
	if err != nil {
		t.Fatalf("TestInit failed: %v", err)
	}

	err = os.Mkdir(searchRoot, DEFAULT_PERM)
	if err != nil {
		t.Fatalf("TestInit failed: %v", err)
	}
}

// TestDirExists checks to make sure none of the specified test directories
// exist yet.
func TestDirExists(t *testing.T) {
	for _, v := range dirs {
		exists, info := DirExists(v)
		if exists {
			t.Fatalf("DirExists %v expected false, got true", v)
		}

		if info != nil {
			t.Fatalf("DirExists %v expected nil FileInfo", v)
		}

		log.Printf("!DirExists %v: passed", v)
	}
}

// TestFileExists checks to make sure none of the specified test files
// exist yet.
func TestFileExists(t *testing.T) {
	for _, v := range files {
		exists, info := FileExists(v)
		if exists {
			t.Fatalf("FileExists %v expected false, got true", v)
		}

		if info != nil {
			t.Fatalf("FileExists %v expected nil FileInfo", v)
		}

		log.Printf("!FileExists %v: passed", v)
	}
}

// TestMkdir attempts to create the test directory structure.
func TestMkdir(t *testing.T) {
	for _, v := range dirs {
		err := Mkdir(v, DEFAULT_PERM)
		if err != nil {
			t.Fatal(err)
		}
	
		exists, info := DirExists(v)
		if !exists {
			t.Fatalf("TestMkdir %v expected to exist, but doesn't", v)
		}

		if info == nil {
			t.Fatalf("TestMkdir %v expected non-nil FileInfo", v)
		}

		log.Printf("TestMkdir %v: passed", v)
	}
}

// TestAppendCreate attempts to create all test files and append ssome 
// data to each.
func TestAppendCreate(t *testing.T) {
	for _, v := range files {
		file, err := AppendFile(v)
		if err != nil {
			t.Fatalf("TestAppendCreate %v failed: %v", v, err)
		}
		defer file.Close()

		txt        := "test"
		count, err := file.WriteString(txt)
		if err != nil {
			t.Fatalf("TestAppendCreate %v failed: %v", v, err)
		}

		if len(txt) != count {
			t.Fatalf(
				"TestAppendCreate %v count wrong. Expected: %v, Observed: %v",
				v,
				len(txt),
				count,
			)
		}
	
		log.Printf("TestAppendCreate %v: passed", v)
	}

	for _, v := range badFiles {
		file, err := AppendFile(v)
		if err == nil {
			file.Close()
			t.Fatalf("TestAppendCreate (badFile) %v failed", v)
		}
	
		log.Printf("TestAppendCreate (badFile) %v: passed", v)
	}
}

// TestAppend attempts to open all previously created files and appends 
// more data to each.
func TestAppend(t *testing.T) {
	for _, v := range files {
		file, err := AppendFile(v)
		if err != nil {
			t.Fatalf("TestAppend %v failed: %v", v, err)
		}
		defer file.Close()

		txt        := "123"
		count, err := file.WriteString(txt)
		if err != nil {
			t.Fatalf("TestAppend %v failed: %v", v, err)
		}

		if len(txt) != count {
			t.Fatalf(
				"TestAppend %v count wrong. Expected: %v, Observed: %v",
				v,
				len(txt),
				count,
			)
		}
	
		log.Printf("TestAppend %v: passed", v)
	}

	for _, v := range badFiles {
		file, err := AppendFile(v)
		if err == nil {
			file.Close()
			t.Fatalf("TestAppend (badFile) %v failed", v)
		}
	
		log.Printf("TestAppend (badFile) %v: passed", v)
	}
}

// TestOpenFile attempts to open all previously created files and ensures
// the correct contents are contained within each.
func TestOpenFile(t *testing.T) {
	expected := "test123"

	for _, v := range files {
		file, err := OpenFile(v)
		if err != nil {
			t.Fatalf("OpenFile %v failed: %v", v, err)
		}
		defer file.Close()

		buffer     := make([]byte, 25)
		count, err := file.Read(buffer)
		if err != nil {
			t.Fatalf("OpenFile %v failed: %v", v, err)
		}

		txt := string(buffer)
		for i := range expected {
			if expected[i] != txt[i] {
				t.Fatalf(
					"OpenFile %v failed: Expected: %v, Observed: %v",
					v,
					expected,
					txt,
				)		
			}
		}

		if len(expected) != count {
			t.Fatalf(
				"OpenFile %v count wrong. Expected: %v, Observed: %v",
				v,
				len(expected),
				count,
			)
		}
	
		log.Printf("OpenFile %v: passed", v)
	}

	for _, v := range badFiles {
		file, err := OpenFile(v)
		if err == nil {
			file.Close()
			t.Fatalf("OpenFile (badFile) %v failed", v)
		}
	
		log.Printf("OpenFile (badFile) %v: passed", v)
	}
}

// TestSearchDirs performs a few pattern matching searches of directories
// in the test directory structure.
func TestSearchDirs(t *testing.T) {
	doSearch(DirsOnly, 14, "*", t)     // 13 dirs + root
	doSearch(DirsOnly, 1,  "tests", t) // root dir
	doSearch(DirsOnly, 0,  "bogus", t) // none exist
	doSearch(DirsOnly, 12, "d*", t)
	doSearch(DirsOnly, 4,  "d2*", t)
	doSearch(DirsOnly, 1,  "c*", t)
	doSearch(DirsOnly, 2,  "d1B*", t)
	doSearch(DirsOnly, 1,  "d1A*", t)
}

// TestSearchFiles performs a few pattern matching searches of files
// in the test directory structure.
func TestSearchFiles(t *testing.T) {
	doSearch(FilesOnly, len(files), "*",        t)
	doSearch(FilesOnly, 4,          "*d*",      t)
	doSearch(FilesOnly, len(files), "*.f",      t)
	doSearch(FilesOnly, 9,          "file*.f",  t)
	doSearch(FilesOnly, 10,         "*file*",   t)
	doSearch(FilesOnly, 5,          "file1*.f", t)
	doSearch(FilesOnly, 2,          "file2.f",  t)
}

// TestSearch performs a few pattern matching searches of files and directories
// in the test directory structure.
func TestSearch(t *testing.T) {
	doSearch(All, 14 + len(files), "*",       t)
	doSearch(All, 16,              "*d*",     t)
	doSearch(All, len(files),      "*.f",     t)
	doSearch(All, 0,               "bogus",   t)
	doSearch(All, 2,               "*tests*", t)
	doSearch(All, 3,               "*fd*",    t)
	doSearch(All, 2,               "*fd.f",   t)
}

// TestEnd cleans up the temporary test directory structure after all tests
// have run.
func TestEnd(t *testing.T) {
	err := os.RemoveAll(searchRoot)
	if err != nil {
		t.Fatalf("TestEnd failed: %v", err)
	}
}

// doSearch is a generalized search routine.
func doSearch(searchType, expectedCount int, searchFilter string, t *testing.T) {
	var testName string
	var matches  int

	switch searchType {
	case All:
		testName = "TestSearch"
		go dir.Search(searchRoot, searchFilter)
	case DirsOnly:
		testName = "TestSearchDirs"
		go dir.SearchDirs(searchRoot, searchFilter)
	case FilesOnly:
		testName = "TestSearchFiles"
		go dir.SearchFiles(searchRoot, searchFilter)
	}

	searching := true

	for searching {
		select {
		case <-dir.DoneChan:
			searching = false
		case err := <-dir.ErrChan:
			t.Fatalf(
				"%v %v failed: %v", 
				testName, 
				searchFilter, 
				err,
			)
		case <-dir.FileChan:
			matches++
		}
	}

	if matches != expectedCount {
		t.Fatalf(
			"%v %v failed: match(%v) != %v", 
			testName,
			searchFilter,
			matches, 
			expectedCount,
		)
	}

	log.Printf(
		"%v %v: passed match(%v)", 
		testName, 
		searchFilter, 
		matches,
	)
}
