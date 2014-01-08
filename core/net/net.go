//  ---------------------------------------------------------------------------
//
//  net.go
//
//  Copyright (c) 2013, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------


package net

var netId uint64


type RawMsg struct {
	cli  *remoteCli
	data []byte
}
