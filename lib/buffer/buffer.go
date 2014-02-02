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

// Stdlib imports.
import (
	"errors"
)

// Data type sizes.
const (
	BYTE_SIZE   = 1
	UINT8_SIZE  = 1
	UINT16_SIZE = 2
	UINT32_SIZE = 4
	UINT64_SIZE = 8
)

// Common errors.
var (
	errExceedBounds = errors.New(
		"Index exceeds capacity of slice/array",
	)
)

// LenByte returns the encoded length of a byte value.
func LenByte() int {
	return BYTE_SIZE
}

// LenString returns the encoded length of a given string, including 
// the extra bytes required to encode its size.
func LenString(val string) int {
	return len(val) + LenUint32()
}

// LenUint8 returns the lengh of a uint32 value.
func LenUint8() int {
	return UINT8_SIZE
}

// LenUint16 returns the lengh of a uint32 value.
func LenUint16() int {
	return UINT16_SIZE
}

// LenUint32 returns the lengh of a uint32 value.
func LenUint32() int {
	return UINT32_SIZE
}

// LenUint64 returns the encoded length of a uint64 value.
func LenUint64() int {
	return UINT64_SIZE
}

// ReadByte reads 1 byte value from the given buffer, starting at
// the given offset, and increments teh supplied cursor by the length
// of the encoded value.
func ReadByte(buffer []byte, cursor *int) (byte, error) {
	if *cursor + BYTE_SIZE > len(buffer) {
		*cursor = len(buffer)
		return 0, errExceedBounds
	}

	val := buffer[*cursor]
	*cursor++

	return val, nil
}

// ReadString reads 1 string value from the given buffer, starting at
// the given offset, and increments teh supplied cursor by the length
// of the encoded value.
func ReadString(buffer []byte, cursor *int) (string, error) {
	size, err := ReadUint32(buffer, cursor)
	if err != nil {
		return "", err
	}

	if *cursor + int(size) > len(buffer) {
		*cursor = len(buffer)
		return "", errExceedBounds
	}

	val     := string(buffer[*cursor:*cursor + int(size)])
	*cursor += len(val)

	return val, nil
}

// ReadUint32 reads 1 uint32 value from the given buffer, starting at
// the given offset, and increments teh supplied cursor by the length
// of the encoded value.
func ReadUint32(buffer []byte, cursor *int) (uint32, error) {
	if *cursor + UINT32_SIZE > len(buffer) {
		*cursor = len(buffer)
		return 0, errExceedBounds
	}

	val := 
		uint32(buffer[*cursor])     << 24 | 
		uint32(buffer[*cursor + 1]) << 16 |
		uint32(buffer[*cursor + 2]) <<  8 |
		uint32(buffer[*cursor + 3])

	*cursor += UINT32_SIZE

	return val, nil
}

// ReadUint64 reads 1 uint64 value from the given buffer, starting at
// the given offset, and increments teh supplied cursor by the length
// of the encoded value.
func ReadUint64(buffer []byte, cursor *int) (uint64, error) {
	if *cursor + UINT64_SIZE > len(buffer) {
		*cursor = len(buffer)
		return 0, errExceedBounds
	}

	val := 
		uint64(buffer[*cursor])     << 56 | 
		uint64(buffer[*cursor + 1]) << 48 |
		uint64(buffer[*cursor + 2]) << 40 |
		uint64(buffer[*cursor + 3]) << 32 |
		uint64(buffer[*cursor + 4]) << 24 | 
		uint64(buffer[*cursor + 5]) << 16 |
		uint64(buffer[*cursor + 6]) <<  8 |
		uint64(buffer[*cursor + 7])

	*cursor += UINT64_SIZE

	return val, nil
}

// WriteByte writes the given byte value into a buffer at the given
// offset and increments the supplied cursor by the length of the encoded
// value.
func WriteByte(val byte, buffer []byte, cursor *int) {
	buffer[*cursor] = val
	*cursor++
}

// WriteString writes the given string value into a buffer starting at
// the given offset, and increments the supplied counter by the length 
// of that encoded string.
func WriteString(val string, buffer []byte, cursor *int) {
	WriteUint32(uint32(len(val)), buffer, cursor)

	count   := copy(buffer[*cursor:], []byte(val))
	*cursor += count
}

// WriteUint32 writes the given uint32 value into a buffer at the given
// offset and increments the supplied cursor by the lengh of the encoded
// value.
func WriteUint32(val uint32, buffer []byte, cursor *int) {
	buffer[*cursor] = byte(val >> 24); *cursor++
	buffer[*cursor] = byte(val >> 16); *cursor++
	buffer[*cursor] = byte(val >> 8);  *cursor++
	buffer[*cursor] = byte(val);       *cursor++
}

// WriteUint64 writes the given uint64 value into a buffer at the given
// offset and increments teh supplied cursor by the length of the encoded
// value.
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
