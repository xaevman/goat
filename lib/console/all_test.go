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

package console

// Stdlib imports.
import (
	"testing"
)

// TestConsole fires up the console system, and writes some lines
// in a few different styles. If it doesn't crash, it passes!
func TestConsole(t *testing.T) {
	WriteLine("")
	WriteLine("Starting test")
	WriteLine("")

	// red on default style
	style1 := Style {
		FG_RED,
		0,
		true,
	}

	// white on green style
	style2 := Style {
		FG_WHITE,
		BG_GREEN,
		false,
	}

	// bold yellow on black style
	style3 := Style {
		FG_YELLOW,
		BG_BLACK,
		true,
	}

	// red on blue style
	style4 := Style {
		FG_RED,
		BG_BLUE,
		false,
	}

	WriteLineFmt("Bold Red on Default", style1)
	WriteLineFmt("White on Green, yah", style2)
	WriteLineFmt("Bold Yellow on Black, yup", style3)
	WriteLineFmt("Red on blue? ouch", style4)
	WriteLine("Back to normal")

	WriteLine("")
}
