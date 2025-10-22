package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// BuildFixed builds bytes for a fixed-size value with a unit separator
func BuildFixed(size int, value uint64) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write the value based on size
	switch size {
	case 1:
		if err := binary.Write(buf, binary.LittleEndian, uint8(value)); err != nil {
			return nil, fmt.Errorf("failed to write 1-byte value: %w", err)
		}
	case 2:
		if err := binary.Write(buf, binary.LittleEndian, uint16(value)); err != nil {
			return nil, fmt.Errorf("failed to write 2-byte value: %w", err)
		}
	case 4:
		if err := binary.Write(buf, binary.LittleEndian, uint32(value)); err != nil {
			return nil, fmt.Errorf("failed to write 4-byte value: %w", err)
		}
	case 8:
		if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
			return nil, fmt.Errorf("failed to write 8-byte value: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid size: %d (must be 1, 2, 4, or 8)", size)
	}

	// Add unit separator
	buf.WriteByte(UnitSeparator)

	return buf.Bytes(), nil
}

// BuildVariable builds bytes for a variable-length string with size prefix and separators
// Format: [StringLength(4 bytes)][UnitSeparator][StringContent][UnitSeparator]
// Example: BuildVariable("banana") returns [06 00][1F][62 61 6E 61 6E 61][1F]
func BuildVariable(value string) ([]byte, error) {
	data := []byte(value)
	length := uint64(len(data))

	// Build length bytes using BuildFixed (2 bytes + separator)
	lengthBytes, err := BuildFixed(2, length)
	if err != nil {
		return nil, fmt.Errorf("failed to build length: %w", err)
	}

	buf := new(bytes.Buffer)

	// Add length part [06 00][1F]
	buf.Write(lengthBytes)

	// Add content [62 61 6E 61 6E 61]
	buf.Write(data)

	// Add separator after content [1F]
	buf.WriteByte(UnitSeparator)

	return buf.Bytes(), nil
}
