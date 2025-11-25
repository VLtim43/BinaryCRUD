package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EntryInfo represents an entry found in the binary file
type EntryInfo struct {
	Data     []byte // The raw entry data (without record separator)
	Position int64  // File offset where this entry starts
}

// SplitFileIntoEntries reads a binary file and splits it into individual entries
// Returns a slice of EntryInfo containing the raw data and file position for each entry
// Format: [recordLength(2)][record data...]
func SplitFileIntoEntries(filePath string) ([]EntryInfo, error) {
	// Read the entire file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if file is large enough to contain a header
	if len(fileData) < HeaderSize {
		return []EntryInfo{}, nil
	}

	entries := make([]EntryInfo, 0)
	offset := HeaderSize

	// Read records using length-prefixed format
	for offset < len(fileData) {
		// Check if we have enough bytes for the length field
		if offset+RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, newOffset, err := ReadFixedNumber(RecordLengthSize, fileData, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read record length at offset %d: %w", offset, err)
		}

		// Check if we have enough bytes for the complete record
		if newOffset+int(recordLength) > len(fileData) {
			return nil, fmt.Errorf("incomplete record at offset %d: expected %d bytes, only %d available",
				newOffset, recordLength, len(fileData)-newOffset)
		}

		// Extract the record data (without the length prefix)
		recordData := fileData[newOffset : newOffset+int(recordLength)]

		entries = append(entries, EntryInfo{
			Data:     recordData,
			Position: int64(newOffset), // Position points to start of record data (after length)
		})

		// Move to next record
		offset = newOffset + int(recordLength)
	}

	return entries, nil
}

// EnsureFileExists creates a binary file with an empty header if it doesn't exist
// This is a common pattern used by all DAOs
func EnsureFileExists(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Create the file
	file, err := CreateFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Extract base filename from path for the header (without extension)
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// Write empty header with fileName (0 entities, 0 tombstones, nextId=0)
	header, err := WriteHeader(fileName, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}
