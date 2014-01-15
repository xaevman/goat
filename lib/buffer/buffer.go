//  ---------------------------------------------------------------------------
//
//  buffer.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package buffer


func ReadString(buffer []byte, cursor *int) string {
	size    := ReadUint64(buffer, cursor)
	val     := string(buffer[*cursor:*cursor + int(size)])
	*cursor += len(val)

	return val
}

func ReadUint32(buffer []byte, cursor *int) uint32 {
	val := 
		uint32(buffer[*cursor])     << 24 | 
		uint32(buffer[*cursor + 1]) << 16 |
		uint32(buffer[*cursor + 2]) <<  8 |
		uint32(buffer[*cursor + 3])

	*cursor += 4

	return val
}

func ReadUint64(buffer []byte, cursor *int) uint64 {
	val := 
		uint64(buffer[*cursor])     << 56 | 
		uint64(buffer[*cursor + 1]) << 48 |
		uint64(buffer[*cursor + 2]) << 40 |
		uint64(buffer[*cursor + 3]) << 32 |
		uint64(buffer[*cursor + 4]) << 24 | 
		uint64(buffer[*cursor + 5]) << 16 |
		uint64(buffer[*cursor + 6]) <<  8 |
		uint64(buffer[*cursor + 7])

	*cursor += 8

	return val
}

func WriteString(val string, buffer []byte, cursor *int) {
	WriteUint64(uint64(len(val)), buffer, cursor)

	count   := copy(buffer[*cursor:], []byte(val))
	*cursor += count
}

func WriteUint32(val uint32, buffer []byte, cursor *int) {
	buffer[*cursor] = byte(val >> 24); *cursor++
	buffer[*cursor] = byte(val >> 16); *cursor++
	buffer[*cursor] = byte(val >> 8);  *cursor++
	buffer[*cursor] = byte(val);       *cursor++
}

func WriteUint64(val uint64, buffer []byte, cursor *int) {
	buffer[*cursor] = byte(val >> 56); *cursor++
	buffer[*cursor] = byte(val >> 48); *cursor++
	buffer[*cursor] = byte(val >> 40); *cursor++
	buffer[*cursor] = byte(val >> 32); *cursor++
	buffer[*cursor] = byte(val >> 24); *cursor++
	buffer[*cursor] = byte(val >> 16); *cursor++
	buffer[*cursor] = byte(val >> 8);  *cursor++
	buffer[*cursor] = byte(val);       *cursor++
}
