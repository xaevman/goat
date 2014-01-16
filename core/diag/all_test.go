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

package diag

import (
	"errors"
	"log"
	"testing"
)

var e = errors.New("This is a test error")

// TestDiag creates diag objects and formats them as strings and json.
// If the process doesn't crash itself, the test passes!
func TestDiag(t *testing.T) {
	str := AsString(e)
	log.Println("AsString ****************************")
	log.Println(str)

	log.Println()
	log.Println()

	json := AsJson(e)
	log.Println("AsJson ******************************")
	log.Println(json)
}
