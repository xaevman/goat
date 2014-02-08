//  ---------------------------------------------------------------------------
//
//  access.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package net

// NoSecurity implements AccessProvider in a way which always returns
// maximum privileges (255).
type NoSecurity struct {}

// Authorize always returns 255, nil.
func (this *NoSecurity) Authorize(con Connection) (byte, error) {
    return 255, nil
}

// Unused.
func (this *NoSecurity) Close() {}

// Unused.
func (this *NoSecurity) Init(proto *Protocol) {}
