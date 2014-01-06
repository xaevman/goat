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

package strutil

import(
	"log"
	"testing"
)

// Test data for DelimToStrArray
var delimTests = []string {
	"val1,val2,val3,val4",
	"val1, val2, val3, val4,",
	"  val1  , ,  val2  , val3 ,, val4     ",
	"val1\tval2\t\tval3\tval4",
	"val1 \tval2 \t  \t  val3 	\t val4",
	"val1\n\n\n\n\nval2\nval3\nval4",
	"val1 \n val2 \n val3 \n val4 \n",
	"val123val223val323val4232323",
}

// Test data for StrArrayToX functions
var arrayTest = []string {
	"London",
	"bridge is falling",
	"\tdown ",
	"\t   falling down",
	" falling down. ",
	"London bridge is falling down\n\n\n",
	"my",
	"fair\t",
	"lady!!",
}

// TestDelimToStrArray tests splitting on various types of separators
// and with varrying amounts and types of extra separators and whitespace
// interspersed with real data.
func TestDelimToStrArray(t *testing.T) {
	delimTest(",",  "commas",   delimTests[:3],  t)
	delimTest("\t", "tabs",     delimTests[3:5], t)
	delimTest("\n", "newlines", delimTests[5:7], t)
	delimTest("23", "chars",    delimTests[7:],  t)
}

// TestStrArrayToLines passes some example data to StrArrayToCsv and makes
// sure that valid CSV format is returned.
func TestStrArrayToCsv(t *testing.T) {
	expect := 
		"London, bridge is falling, down, falling down, falling down., " +
		"London bridge is falling down, my, fair, lady!!"

	result := StrArrayToCsv(arrayTest)
	if !StrEq(expect, result) {
		t.Fatalf(
			"TestStrArrayToCsv failed: expect (%v), result (%v)",
			expect,
			result,
		)
	}

	log.Printf("TestStrArrayToCsv passed")
}

// TestStrArrayToLines passes some example data to StrArrayToLines and validates
// the result.
func TestStrArrayToLines(t *testing.T) {
	expect := 
		"London\nbridge is falling\ndown\nfalling down\nfalling down.\n" +
		"London bridge is falling down\nmy\nfair\nlady!!"

	result := StrArrayToLines(arrayTest)
	if !StrEq(expect, result) {
		t.Fatalf(
			"StrArrayToLines failed: expect (%v), result (%v)",
			expect,
			result,
		)
	}

	log.Printf("StrArrayToLines passed")

}

// delimTest is a wrapper for DelimToStrArray for performing tests with
// different delimiters, on different parts of the delimTests array.
func delimTest(sep, testName string, testData []string, t *testing.T) {
	var result []string
	expected := []string {
		"val1", 
		"val2", 
		"val3", 
		"val4",
	}

	for i := range testData {
		result = DelimToStrArray(testData[i], sep)
		if len(result) != len(expected) {
			t.Fatalf(
				"TestDelimToStrArray %v[%v] failed: lr(%v), le(%v)",
				testName,
				i,
				len(result),
				len(expected),
			)
		}
		
		for i := range result {
			if !StrEq(result[i], expected[i]) {
				t.Fatalf(
					"TestDelimtoStrArray %v[%v] failed: %v != %v",
					testName,
					i,
					result[i],
					expected[i],
				)
			}
		}

		log.Printf("TestDelimToStrArray %v[%v]: passed", testName, i)
	}
}
