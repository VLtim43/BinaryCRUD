package utils

import (
	"encoding/binary"
	"fmt"
	"os"
)

// ReadVariable reads a variable-length string from the file
// Format: [Length(2)][UnitSeparator][Content][UnitSeparator]
func ReadVariable(file *os.File) (string, error) {
	// Read length (2 bytes, little-endian)
	var length uint16
	if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	// Read and verify unit separator after length
	sep := make([]byte, 1)
	if _, err := file.Read(sep); err != nil {
		return "", fmt.Errorf("failed to read unit separator after length: %w", err)
	}
	if sep[0] != UnitSeparator {
		return "", fmt.Errorf("invalid unit separator after length: expected 0x1F, got 0x%X", sep[0])
	}

	// Read string content
	data := make([]byte, length)
	if _, err := file.Read(data); err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}

	// Read and verify unit separator after content
	if _, err := file.Read(sep); err != nil {
		return "", fmt.Errorf("failed to read unit separator after content: %w", err)
	}
	if sep[0] != UnitSeparator {
		return "", fmt.Errorf("invalid unit separator after content: expected 0x1F, got 0x%X", sep[0])
	}

	result := string(data)

	// Debug print the read content
	DebugPrint("Read", result)

	return result, nil
}
