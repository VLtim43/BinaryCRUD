package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// Unicode separators for unit and record separation
const (
	UnitSeparator   = '\x1F' // U+001F - Unit Separator
	RecordSeparator = '\x1E' // U+001E - Record Separator
)

// BinaryFileHeader represents the common header structure for all binary files
type BinaryFileHeader struct {
	EntryCount     uint32 // Number of entries in the file
	TombstoneCount uint32 // Number of deleted entries (tombstones)
}

// HeaderSize is the size of the binary file header in bytes
// Format: [EntryCount(4)][UnitSeparator(1)][TombstoneCount(4)][RecordSeparator(1)]
const HeaderSize = 10 // 4 + 1 + 4 + 1 = 10 bytes

// InitializeBinaryFile creates the data directory and initializes a binary file with header
func InitializeBinaryFile(filePath string) error {
	// Create data directory if it doesn't exist
	dataDir := filepath.Dir(filePath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, no need to initialize
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Create new file with initialized header
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write initial header (all zeros)
	header := BinaryFileHeader{
		EntryCount:     0,
		TombstoneCount: 0,
	}

	if err := WriteHeader(file, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// WriteHeader writes the header to the beginning of the file
// Format: [EntryCount(4)][UnitSeparator(1)][TombstoneCount(4)][RecordSeparator(1)]
func WriteHeader(file *os.File, header BinaryFileHeader) error {
	// Seek to beginning of file
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write EntryCount (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, header.EntryCount); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}

	// Write Unit Separator
	if _, err := file.Write([]byte{UnitSeparator}); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}

	// Write TombstoneCount (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, header.TombstoneCount); err != nil {
		return fmt.Errorf("failed to write tombstone count: %w", err)
	}

	// Write Record Separator
	if _, err := file.Write([]byte{RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	return nil
}

// ReadHeader reads the header from the beginning of the file
// Format: [EntryCount(4)][UnitSeparator(1)][TombstoneCount(4)][RecordSeparator(1)]
func ReadHeader(file *os.File) (BinaryFileHeader, error) {
	var header BinaryFileHeader

	// Seek to beginning of file
	if _, err := file.Seek(0, 0); err != nil {
		return header, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Read EntryCount (4 bytes, little-endian)
	if err := binary.Read(file, binary.LittleEndian, &header.EntryCount); err != nil {
		return header, fmt.Errorf("failed to read entry count: %w", err)
	}

	// Read and verify Unit Separator
	sep := make([]byte, 1)
	if _, err := file.Read(sep); err != nil {
		return header, fmt.Errorf("failed to read unit separator: %w", err)
	}
	if sep[0] != UnitSeparator {
		return header, fmt.Errorf("invalid unit separator: expected 0x1F, got 0x%X", sep[0])
	}

	// Read TombstoneCount (4 bytes, little-endian)
	if err := binary.Read(file, binary.LittleEndian, &header.TombstoneCount); err != nil {
		return header, fmt.Errorf("failed to read tombstone count: %w", err)
	}

	// Read and verify Record Separator
	if _, err := file.Read(sep); err != nil {
		return header, fmt.Errorf("failed to read record separator: %w", err)
	}
	if sep[0] != RecordSeparator {
		return header, fmt.Errorf("invalid record separator: expected 0x1E, got 0x%X", sep[0])
	}

	return header, nil
}

// UpdateHeader updates the header in the file with new values
func UpdateHeader(filePath string, header BinaryFileHeader) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return WriteHeader(file, header)
}

// IncrementEntryCount increments the entry count in the header by 1
func IncrementEntryCount(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	header, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	header.EntryCount++

	return WriteHeader(file, header)
}

// IncrementTombstoneCount increments the tombstone count in the header by 1
func IncrementTombstoneCount(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	header, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	header.TombstoneCount++

	return WriteHeader(file, header)
}

// GetHeaderInfo returns the header information from a binary file
func GetHeaderInfo(filePath string) (BinaryFileHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return BinaryFileHeader{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ReadHeader(file)
}
