//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Written by Jared Chavez (2014-01-01)
//  Owned by Jared Chavez <xaevman@gmail.com>
//
//  Copyright (c) 2014 Jared Chavez
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

	json := FmtDiagJson(e)
	log.Println("AsJson ******************************")
	log.Println(json)
}
