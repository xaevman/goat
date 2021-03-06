//  ---------------------------------------------------------------------------
//
//  string.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

// Package str contains helper functions for commonly useful string
// operations.
package str

// Stdlib imports.
import (
    "bytes"
    "strconv"
    "strings"
)


// DelimToStrArray takes a data string, delimited by a specified separator,
// and splits the string by that separator, also trimming whitespace and
// dropping any resulting empty strings.
func DelimToStrArray(data, sep string) []string {
    temp := make([]string, 0)
    vals := strings.Split(data, sep) 

    for i := range vals {
        cleanVal := strings.TrimSpace(vals[i])
        if len(cleanVal) > 0 {
            temp = append(temp, cleanVal)
        }
    }

    return temp
}

// IntArrayToList takes an array of ints and transforms them into a 
// single string, delimited by a specified separator.
func IntArrayToList(items []int, sep string) string {
    var buffer bytes.Buffer

    for i, v := range items {
        buffer.WriteString(strings.TrimSpace(strconv.Itoa(v)))
        if i < len(items) - 1 {
            buffer.WriteString(sep)
        }
    }

    return buffer.String()
}

// Int64ArrayToList takes an array of int64s and transforms them into a 
// single string, delimited by a specified separator.
func Int64ArrayToList(items []int64, sep string) string {
    var buffer bytes.Buffer

    for i, v := range items {
        buffer.WriteString(strings.TrimSpace(strconv.FormatInt(v, 10)))
        if i < len(items) - 1 {
            buffer.WriteString(sep)
        }
    }

    return buffer.String()
}

// StrArrayToCsv is a helper function that calls StrArrayToList 
// using ", " as a separator.
func StrArrayToCsv(items []string) string {
    return StrArrayToList(items, ", ")
}

// StrArrayToLines is a helper function that calls StrArrayToList
// using "\n" as a separator.
func StrArrayToLines(items []string) string {
    return StrArrayToList(items, "\n")
}

// StrArrayToList takes an array of strings and transforms them into a 
// single string, delimited by a specified separator.
func StrArrayToList(items []string, sep string) string {
    var buffer bytes.Buffer

    for i, v := range items {
        buffer.WriteString(strings.TrimSpace(v))
        if i < len(items) - 1 {
            buffer.WriteString(sep)
        }
    }

    return buffer.String()
}

// StrEq compares two strings, character for character, and returns 
// true if they are the same.
func StrEq(s1, s2 string) bool {
    if len(s1) != len(s2) {
        return false
    }

    for i := range s1 {
        if s1[i] != s2[i] {
            return false
        }
    }

    return true
}
