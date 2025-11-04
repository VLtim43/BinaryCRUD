package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	UnitSeparator   = '\x1F' // U+001F - Unit Separator
	RecordSeparator = '\x1E' // U+001E - Record Separator
)

type BinaryFileHeader struct {
	EntryCount     uint32
	TombstoneCount uint32
	NextID         uint32
}

const HeaderSize = 14

func InitializeBinaryFile(filePath string) error {
	dataDir := filepath.Dir(filePath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if _, err := os.Stat(filePath); err == nil {
		// File exists, no need to initialize
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write initial header (all zeros, NextID starts at 0)
	initialHeader := BinaryFileHeader{
		EntryCount:     0,
		TombstoneCount: 0,
		NextID:         0,
	}

	if err := WriteHeader(file, initialHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

func WriteHeader(file *os.File, entryCount uint32, tombstoneCount uint32, nextID uint32) error {
	// Seek to beginning of file
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write EntryCount (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, entryCount); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}

	// Write Unit Separator
	if _, err := file.Write([]byte{UnitSeparator}); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}

	// Write TombstoneCount (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, tombstoneCount); err != nil {
		return fmt.Errorf("failed to write tombstone count: %w", err)
	}

	// Write Unit Separator
	if _, err := file.Write([]byte{UnitSeparator}); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}

	// Write NextID (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, nextID); err != nil {
		return fmt.Errorf("failed to write next ID: %w", err)
	}

	// Write Record Separator
	if _, err := file.Write([]byte{RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	return nil
}

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

	// Read and verify Unit Separator
	if _, err := file.Read(sep); err != nil {
		return header, fmt.Errorf("failed to read unit separator after tombstone: %w", err)
	}
	if sep[0] != UnitSeparator {
		return header, fmt.Errorf("invalid unit separator after tombstone: expected 0x1F, got 0x%X", sep[0])
	}

	// Read NextID (4 bytes, little-endian)
	if err := binary.Read(file, binary.LittleEndian, &header.NextID); err != nil {
		return header, fmt.Errorf("failed to read next ID: %w", err)
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

func UpdateHeader(filePath string, header BinaryFileHeader) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return WriteHeader(file, header)
}

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

func GetNextIDAndIncrement(filePath string) (uint32, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	header, err := ReadHeader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	// Store the current NextID to return
	currentID := header.NextID

	// Increment NextID for next time
	header.NextID++

	// Write updated header
	if err := WriteHeader(file, header); err != nil {
		return 0, fmt.Errorf("failed to write header: %w", err)
	}

	return currentID, nil
}

func GetHeaderInfo(filePath string) (BinaryFileHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return BinaryFileHeader{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ReadHeader(file)
}

func OpenBinaryFile(filePath string) (*os.File, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func OpenFileForWrite(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}
