//  ---------------------------------------------------------------------------
//
//  file.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package fs

import (
	"os"
)

// Default permissions for simple file operations (u:rwx, g:rx, o: ).
const DEFAULT_PERM = 0750


// AppendFile creates or opens the specified path with append+write-only
// flags and DEFAULT_PERM permissions.
func AppendFile(path string) (*os.File, error) {
	return AppendFilePerm(path, DEFAULT_PERM)
}

// AppendFilePerm creates or opens the sepcified path with append+write-only
// flags at the requested permissions level.
func AppendFilePerm(path string, perm os.FileMode) (*os.File, error) {
	if exists, _ := FileExists(path); exists {
		return os.OpenFile(
			path,
			os.O_APPEND|os.O_WRONLY,
			perm,
		)
	}

	return os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		perm,
	)
}

// FileExists tests to see if a given file exists on disk and returns
// true, along with a relevant os.FileInfo object if so.
func FileExists(path string) (bool, os.FileInfo) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err), fileInfo
	}

	return true, fileInfo
}

// GetFileSize stats a file on disk and returns its length. If no file exists
// then zero is returned.
func GetFileSize(path string) int64 {
	exist, info := FileExists(path)
	if !exist {
		return 0
	}

	return info.Size()
}

// OpenFile creates or opens the specified path with the read-only
// flag and DEFAULT_PERM permissions.
func OpenFile(path string) (*os.File, error) {
	return OpenFilePerm(path, DEFAULT_PERM)
}

// OpenFilePerm creates or opens the specified path with the read-only 
// flag at the requested permissions level.
func OpenFilePerm(path string, perm os.FileMode) (*os.File, error) {
	if exists, _ := FileExists(path); exists {
		return os.OpenFile(
			path,
			os.O_RDONLY,
			perm,
		)
	}

	return os.OpenFile(
		path,
		os.O_CREATE|os.O_EXCL|os.O_RDONLY,
		perm,
	)
}
