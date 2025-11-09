package utils

import (
	"encoding/binary"
	"fmt"
)

// ReadFixedNumber reads a fixed-size number from a byte slice at the given offset
func ReadFixedNumber(size int, data []byte, offset int) (uint64, int, error) {
	if size <= 0 {
		return 0, offset, fmt.Errorf("size must be positive")
	}
	if size > 8 {
		return 0, offset, fmt.Errorf("size cannot exceed 8 bytes for uint64")
	}

	if offset+size > len(data) {
		return 0, offset, fmt.Errorf("not enough data: need %d bytes at offset %d, have %d total", size, offset, len(data))
	}

	// Extract the bytes
	bytes := data[offset : offset+size]

	// Convert bytes to uint64 (big-endian)
	// Pad to 8 bytes if necessary
	padded := make([]byte, 8)
	copy(padded[8-size:], bytes)
	value := binary.BigEndian.Uint64(padded)

	return value, offset + size, nil
}

// ReadFixedString reads a fixed-size string from a byte slice at the given offset
func ReadFixedString(size int, data []byte, offset int) (string, int, error) {
	if size <= 0 {
		return "", offset, fmt.Errorf("size must be positive")
	}

	if offset+size > len(data) {
		return "", offset, fmt.Errorf("not enough data: need %d bytes at offset %d, have %d total", size, offset, len(data))
	}

	// Extract the bytes
	bytes := data[offset : offset+size]

	// Convert to string, trimming leading zeros
	str := string(bytes)
	// Trim left zeros
	i := 0
	for i < len(str) && str[i] == 0 {
		i++
	}

	return str[i:], offset + size, nil
}
