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

package buffer

// Stdlib imports.
import (
    "log"
    "testing"
)

// string test values and their encdoed lengths.
var testStr = map[string]int {
    "": 
        UINT32_SIZE,
    "hello": 
        UINT32_SIZE + 5,
    "test string 1": 
        UINT32_SIZE + 13,
    "this is a really long string separated\nby new lines!\n": 
        UINT32_SIZE + 53,
}

// byte test values.
var testByte = []byte {
    0, 1, 10, 16, 24, 96, 128, 200, 255,
}

// uint32 test values.
var testUint32 = []uint32 {
    0, 1, 16, 550, 65000,
}

// uint64 test values.
var testUint64 = []uint64 {
    0, 1, 16, 550, 65000, 99000000, 1270999888,
}

// TestLengths tests the Len helper functions to make sure they return
// known values.
func TestLengths(t *testing.T) {
    // strings
    for k, v := range testStr {
        val := LenString(k)
        if val != v {
            t.Fatalf("%v returned %v (expected %v)", k, val, v)
        }
    }

    // bytes
    bVal := LenByte()
    if bVal != BYTE_SIZE {
        t.Fatalf("Byte returned %v (expected %v)", bVal, 4)
    }

    // uints
    uval32 := LenUint32()
    if uval32 != UINT32_SIZE {
        t.Fatalf("Uint32 returned %v (expected %v)", uval32, 4)
    }
    uval64 := LenUint64()
    if uval64 != UINT64_SIZE {
        t.Fatalf("Uint64 returned %v (expected %v)", uval64, 8)
    }

    log.Println("TestLengths: passed")
}

// TestRoundTrips takes values, writes them to a byte array, and pulls them 
// back out again, making sure the resulting set of values matches the originals.
func TestRoundTrips(t *testing.T) {
    var totalSize int = 0

    // figure out total buffer size
    for _, v := range testStr {
        totalSize += v
    }

    for _ = range testByte {
        totalSize += BYTE_SIZE
    }

    for _ = range testUint32 {
        totalSize += UINT32_SIZE
    }

    for _ = range testUint64 {
        totalSize += UINT64_SIZE
    }

    cursor := 0
    buffer := make([]byte, totalSize)

    // write all data to buffer
    for k, _ := range testStr {
        WriteString(k, buffer, &cursor)
    }

    for i := range testByte {
        WriteByte(testByte[i], buffer, &cursor)
    }

    for i := range testUint32 {
        WriteUint32(testUint32[i], buffer, &cursor)
    }

    for i := range testUint64 {
        WriteUint64(testUint64[i], buffer, &cursor)
    }

    if cursor != totalSize {
        t.Fatalf(
            "Cursor doesn't match measured size of data: %v vs %v",
            cursor,
            totalSize,
        )
    }

    // retrieve all data from buffer and check round trip values
    cursor = 0

    for k, _ := range testStr {
        val, err := ReadString(buffer, &cursor)
        if err != nil {
            t.Fatal(err)
        }
        if !strEq(val, k) {
            t.Fatalf("Values don't match: %v != %v", val, k)
        }
    }

    for i := range testByte {
        val, err := ReadByte(buffer, &cursor)
        if err != nil {
            t.Fatal(err)
        }
        if val != testByte[i] {
            t.Fatalf("Values don't match: %v != %v", val, testByte[i])
        }
    }

    for i := range testUint32 {
        val, err := ReadUint32(buffer, &cursor)
        if err != nil {
            t.Fatal(err)
        }
        if val != testUint32[i] {
            t.Fatalf("Values don't match: %v != %v", val, testUint32[i])
        }
    }

    for i := range testUint64 {
        val, err := ReadUint64(buffer, &cursor)
        if err != nil {
            t.Fatal(err)
        }
        if val != testUint64[i] {
            t.Fatalf("Values don't match: %v != %v", val, testUint64[i])
        }
    }

    log.Println("TestRoundTrips: passed")
}

// strEq compares two strings, character for character, and returns 
// true if they are the same.
func strEq(s1, s2 string) bool {
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
