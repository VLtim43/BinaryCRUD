package utils

import (
	"fmt"
	"io"
	"os"
)

// FindByIDSequential performs a sequential scan of a binary file to find an entry by ID
// Returns the complete entry data (including ID) and nil error if found
// Returns nil and an error if not found or on file read errors
// Format: [recordLength(2)][ID(2)][tombstone(1)][data...]
func FindByIDSequential(file *os.File, targetID uint64) ([]byte, error) {
	// Seek to the start of the first entry (after header)
	_, err := file.Seek(int64(HeaderSize), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past header: %w", err)
	}

	// Read the rest of the file
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// If file is empty (no entries), return not found
	if len(fileData) == 0 {
		return nil, fmt.Errorf("entry with ID %d not found", targetID)
	}

	// Parse records using length-prefixed format
	offset := 0
	for offset < len(fileData) {
		// Check if we have enough bytes for the length field
		if offset+RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, lengthEnd, err := ReadFixedNumber(RecordLengthSize, fileData, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read record length: %w", err)
		}

		// Check if we have enough bytes for the complete record
		if lengthEnd+int(recordLength) > len(fileData) {
			return nil, fmt.Errorf("incomplete record at offset %d", offset)
		}

		// Extract the record data (without length prefix)
		entryData := fileData[lengthEnd : lengthEnd+int(recordLength)]

		// Each record starts with ID (IDSize bytes)
		if len(entryData) >= IDSize {
			// Read the ID from the entry
			entryID, _, err := ReadFixedNumber(IDSize, entryData, 0)
			if err == nil && entryID == targetID {
				// Found it! Return the complete entry data (including ID)
				return entryData, nil
			}
		}

		// Move to next record
		offset = lengthEnd + int(recordLength)
	}

	// Not found
	return nil, fmt.Errorf("entry with ID %d not found", targetID)
}
