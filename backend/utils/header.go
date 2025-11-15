package utils

import (
	"fmt"
	"os"
)

// WriteHeader creates a header byte slice with entitiesCount, tombstoneCount, and nextId
// Format: [entitiesCount(4)][tombstoneCount(4)][nextId(4)] = 12 bytes
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

	// Build the header: [entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	header := make([]byte, 0, HeaderSize)
	header = append(header, entitiesBytes...)
	header = append(header, tombstoneBytes...)
	header = append(header, nextIdBytes...)

	return header, nil
}

// ReadHeader reads and parses the header from a file
// Returns (entitiesCount, tombstoneCount, nextId, error)
func ReadHeader(file *os.File) (int, int, int, error) {
	// Read the header bytes
	headerBytes := make([]byte, HeaderSize)
	_, err := file.Seek(0, 0) // Start from beginning
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	n, err := file.Read(headerBytes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read header: %w", err)
	}
	if n != HeaderSize {
		return 0, 0, 0, fmt.Errorf("incomplete header: read %d bytes, expected %d", n, HeaderSize)
	}

	// Parse the header: [entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	offset := 0

	// Read entitiesCount
	entitiesCount, offset, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read entitiesCount: %w", err)
	}

	// Read tombstoneCount
	tombstoneCount, offset, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read tombstoneCount: %w", err)
	}

	// Read nextId
	nextId, _, err := ReadFixedNumber(HeaderFieldSize, headerBytes, offset)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read nextId: %w", err)
	}

	return int(entitiesCount), int(tombstoneCount), int(nextId), nil
}

// UpdateHeader updates the header in the file with new values
// This function now syncs the file to disk after writing the header to ensure data integrity
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

	// Force write to disk to ensure header is persisted
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync header to disk: %w", err)
	}

	return nil
}
