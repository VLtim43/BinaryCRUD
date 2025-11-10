package utils

import (
	"fmt"
	"os"
)

// WriteHeader creates a header byte slice with entitiesCount, tombstoneCount, and nextId
// separated by unit separators. Format: [count][0x1F][count][0x1F][count][0x1E]
func WriteHeader(entitiesCount, tombstoneCount, nextId int) ([]byte, error) {
	entitiesBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(entitiesCount))
	if err != nil {
		return nil, fmt.Errorf("failed to write entitiesCount: %w", err)
	}

	tombstoneBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(tombstoneCount))
	if err != nil {
		return nil, fmt.Errorf("failed to write tombstoneCount: %w", err)
	}

	nextIdBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(nextId))
	if err != nil {
		return nil, fmt.Errorf("failed to write nextId: %w", err)
	}

	unitSeparatorBytes, err := WriteVariable(UnitSeparator)
	if err != nil {
		return nil, fmt.Errorf("failed to write unit separator: %w", err)
	}

	recordSeparatorBytes, err := WriteVariable(RecordSeparator)
	if err != nil {
		return nil, fmt.Errorf("failed to write record separator: %w", err)
	}

	// Build the header: [entitiesCount][0x1F][tombstoneCount][0x1F][nextId][0x1E]
	header := make([]byte, 0)
	header = append(header, entitiesBytes...)
	header = append(header, unitSeparatorBytes...)
	header = append(header, tombstoneBytes...)
	header = append(header, unitSeparatorBytes...)
	header = append(header, nextIdBytes...)
	header = append(header, recordSeparatorBytes...)

	return header, nil
}

// ReadHeader reads and parses the header from a file
// Returns (entitiesCount, tombstoneCount, nextId, error)
func ReadHeader(file *os.File) (int, int, int, error) {
	// Calculate header size: 3 fields * 4 bytes each + 2 unit separators + 1 record separator = 15 bytes
	headerSize := (HeaderFieldSize * 3) + 3

	// Read the header bytes
	headerBytes := make([]byte, headerSize)
	_, err := file.Seek(0, 0) // Start from beginning
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	n, err := file.Read(headerBytes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read header: %w", err)
	}
	if n != headerSize {
		return 0, 0, 0, fmt.Errorf("incomplete header: read %d bytes, expected %d", n, headerSize)
	}

	// Parse the header: [entitiesCount][0x1F][tombstoneCount][0x1F][nextId][0x1E]
	offset := 0

	// Read entitiesCount
	entitiesCount, offset, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read entitiesCount: %w", err)
	}

	// Skip unit separator (0x1F = 1 byte)
	offset += 1

	// Read tombstoneCount
	tombstoneCount, offset, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read tombstoneCount: %w", err)
	}

	// Skip unit separator
	offset += 1

	// Read nextId
	nextId, offset, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read nextId: %w", err)
	}

	// Skip record separator (0x1E = 1 byte) - already at end
	// offset += 1

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
	_, err = file.Write(header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}
