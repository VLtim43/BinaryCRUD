package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// BuildFixed builds bytes for a fixed-size value with a unit separator
// size: number of bytes (1, 2, 4, or 8)
// value: the value to encode
// Example: BuildFixed(4, 6) returns [06 00 00 00][1F]
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
// Example: BuildVariable("banana") returns [06 00 00 00][1F][62 61 6E 61 6E 61][1F]
func BuildVariable(value string) ([]byte, error) {
	data := []byte(value)
	length := uint64(len(data))

	// Build length bytes using BuildFixed (4 bytes + separator)
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

// AppendRecord appends pre-built record bytes to the binary file
// recordBytes should contain the complete record data (fields + separators)
// This function will add the final record separator and update the header
func AppendRecord(filePath string, recordBytes []byte) error {
	// Initialize the binary file if it doesn't exist
	if err := InitializeBinaryFile(filePath); err != nil {
		return fmt.Errorf("failed to initialize binary file: %w", err)
	}

	// Open file in read-write mode
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to end of file to append the new record
	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	// Write the record bytes
	if _, err := file.Write(recordBytes); err != nil {
		return fmt.Errorf("failed to write record data: %w", err)
	}

	// Write record separator to mark end of record
	if _, err := file.Write([]byte{RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	// Increment entry count in header
	if err := IncrementEntryCount(filePath); err != nil {
		return fmt.Errorf("failed to increment entry count: %w", err)
	}

	return nil
}
