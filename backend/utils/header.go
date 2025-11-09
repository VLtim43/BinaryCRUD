package utils

import (
	"fmt"
	"os"
)

// WriteHeader creates a header string with entitiesCount, tombstoneCount, and nextId
// separated by unit separators. Format: [count][0x1F][count][0x1F][count]
func WriteHeader(entitiesCount, tombstoneCount, nextId int) (string, error) {
	entitiesHex, err := WriteFixedNumber(HeaderFieldSize, uint64(entitiesCount))
	if err != nil {
		return "", fmt.Errorf("failed to write entitiesCount: %w", err)
	}

	tombstoneHex, err := WriteFixedNumber(HeaderFieldSize, uint64(tombstoneCount))
	if err != nil {
		return "", fmt.Errorf("failed to write tombstoneCount: %w", err)
	}

	nextIdHex, err := WriteFixedNumber(HeaderFieldSize, uint64(nextId))
	if err != nil {
		return "", fmt.Errorf("failed to write nextId: %w", err)
	}

	separatorHex, err := WriteVariable(UnitSeparator)
	if err != nil {
		return "", fmt.Errorf("failed to write separator: %w", err)
	}

	// Build the header: [entitiesCount][0x1F][tombstoneCount][0x1F][nextId]
	header := entitiesHex + separatorHex + tombstoneHex + separatorHex + nextIdHex

	return header, nil
}

// ReadHeader reads and parses the header from a file
// Returns (entitiesCount, tombstoneCount, nextId, error)
func ReadHeader(file *os.File) (int, int, int, error) {
	// Calculate header size: 3 fields * 4 bytes each + 2 separators * 1 byte each = 14 bytes = 28 hex chars
	headerSize := (HeaderFieldSize * 3) + 2 // in bytes
	hexChars := headerSize * 2

	// Read the header bytes
	headerBytes := make([]byte, hexChars)
	_, err := file.Seek(0, 0) // Start from beginning
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	n, err := file.Read(headerBytes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read header: %w", err)
	}
	if n != hexChars {
		return 0, 0, 0, fmt.Errorf("incomplete header: read %d bytes, expected %d", n, hexChars)
	}

	hexString := string(headerBytes)

	// Parse the header: [entitiesCount][0x1F][tombstoneCount][0x1F][nextId]
	offset := 0

	// Read entitiesCount
	entitiesCount, offset, err := ReadFixedNumber(HeaderFieldSize, hexString, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read entitiesCount: %w", err)
	}

	// Skip separator (0x1F = 2 hex chars)
	offset += 2

	// Read tombstoneCount
	tombstoneCount, offset, err := ReadFixedNumber(HeaderFieldSize, hexString, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read tombstoneCount: %w", err)
	}

	// Skip separator
	offset += 2

	// Read nextId
	nextId, _, err := ReadFixedNumber(HeaderFieldSize, hexString, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read nextId: %w", err)
	}

	return int(entitiesCount), int(tombstoneCount), int(nextId), nil
}

// UpdateHeader updates the header in the file with new values
func UpdateHeader(file *os.File, entitiesCount, tombstoneCount, nextId int) error {
	// Generate new header
	header, err := WriteHeader(entitiesCount, tombstoneCount, nextId)
	if err != nil {
		return fmt.Errorf("failed to generate header: %w", err)
	}

	// Seek to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write the header
	_, err = file.WriteString(header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}
