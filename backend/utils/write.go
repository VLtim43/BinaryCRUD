package utils

import (
	"encoding/binary"
	"fmt"
)

// WriteFixedNumber writes a number as binary in a fixed-size field
func WriteFixedNumber(size int, value uint64) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}
	if size > 8 {
		return nil, fmt.Errorf("size cannot exceed 8 bytes for uint64")
	}

	// Check if value fits in the specified number of bytes
	maxValue := uint64(1<<(size*8)) - 1
	if value > maxValue {
		return nil, fmt.Errorf("value %d exceeds maximum for %d bytes (%d)", value, size, maxValue)
	}

	// Convert number to bytes (big-endian)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)

	// Take the last 'size' bytes (right-most bytes)
	data := buf[8-size:]

	return data, nil
}

// WriteFixedString writes a string as ASCII in a fixed-size field
func WriteFixedString(size int, content string) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	contentBytes := []byte(content)

	if len(contentBytes) > size {
		return nil, fmt.Errorf("content length %d exceeds size %d", len(contentBytes), size)
	}

	data := make([]byte, size)
	// left-pad with zeros
	copy(data[size-len(contentBytes):], contentBytes)

	return data, nil
}

// WriteVariable writes a string as ASCII with variable length
func WriteVariable(content string) ([]byte, error) {
	return []byte(content), nil
}
