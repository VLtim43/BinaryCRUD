package utils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// ReadFixedNumber reads a fixed-size number from a hex string at the given offset
func ReadFixedNumber(size int, hexString string, offset int) (uint64, int, error) {
	if size <= 0 {
		return 0, offset, fmt.Errorf("size must be positive")
	}
	if size > 8 {
		return 0, offset, fmt.Errorf("size cannot exceed 8 bytes for uint64")
	}

	// Calculate required hex characters (2 per byte)
	hexChars := size * 2
	if offset+hexChars > len(hexString) {
		return 0, offset, fmt.Errorf("not enough data: need %d chars at offset %d, have %d total", hexChars, offset, len(hexString))
	}

	// Extract the hex substring
	hexSubstr := hexString[offset : offset+hexChars]

	// Decode from hex
	bytes, err := hex.DecodeString(hexSubstr)
	if err != nil {
		return 0, offset, fmt.Errorf("failed to decode hex: %w", err)
	}

	// Convert bytes to uint64 (big-endian)
	// Pad to 8 bytes if necessary
	padded := make([]byte, 8)
	copy(padded[8-size:], bytes)
	value := binary.BigEndian.Uint64(padded)

	return value, offset + hexChars, nil
}

// ReadFixedString reads a fixed-size string from a hex string at the given offset
func ReadFixedString(size int, hexString string, offset int) (string, int, error) {
	if size <= 0 {
		return "", offset, fmt.Errorf("size must be positive")
	}

	// Calculate required hex characters (2 per byte)
	hexChars := size * 2
	if offset+hexChars > len(hexString) {
		return "", offset, fmt.Errorf("not enough data: need %d chars at offset %d, have %d total", hexChars, offset, len(hexString))
	}

	// Extract the hex substring
	hexSubstr := hexString[offset : offset+hexChars]

	// Decode from hex
	bytes, err := hex.DecodeString(hexSubstr)
	if err != nil {
		return "", offset, fmt.Errorf("failed to decode hex: %w", err)
	}

	// Convert to string, trimming leading zeros
	str := string(bytes)
	// Trim left zeros
	i := 0
	for i < len(str) && str[i] == 0 {
		i++
	}

	return str[i:], offset + hexChars, nil
}
