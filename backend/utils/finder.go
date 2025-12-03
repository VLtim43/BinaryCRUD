package utils

import (
	"fmt"
	"io"
	"os"
)

// ReadEntryAtOffset reads a record from a file at the given offset
// The offset should point to the start of the record (at the length prefix)
// Returns the entry data (without length prefix) or nil if read fails
func ReadEntryAtOffset(file *os.File, offset int64) ([]byte, error) {
	// Validate offset is within file bounds
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	if err := ValidateOffset(offset, fileInfo.Size()); err != nil {
		return nil, err
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to offset: %w", err)
	}

	// Read record length (2 bytes)
	lengthBytes := make([]byte, RecordLengthSize)
	n, err := file.Read(lengthBytes)
	if err != nil || n != RecordLengthSize {
		return nil, fmt.Errorf("failed to read record length")
	}

	recordLength, _, err := ReadFixedNumber(RecordLengthSize, lengthBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse record length: %w", err)
	}

	// Validate record length
	if err := ValidateRecordLength(recordLength); err != nil {
		return nil, err
	}

	// Read the record data
	entryData := make([]byte, recordLength)
	n, err = file.Read(entryData)
	if err != nil || n != int(recordLength) {
		return nil, fmt.Errorf("failed to read record data")
	}

	return entryData, nil
}

// FindByIDSequential performs a sequential scan of a binary file to find an entry by ID
// Returns the complete entry data (including ID) and nil error if found
// Returns nil and an error if not found or on file read errors
// Format: [recordLength(2)][ID(2)][tombstone(1)][data...]
func FindByIDSequential(file *os.File, targetID uint64) ([]byte, error) {
	// Get actual header size (variable due to filename)
	headerSize, err := GetHeaderSizeFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get header size: %w", err)
	}

	// Seek to the start of the first entry (after header)
	_, err = file.Seek(int64(headerSize), 0)
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

		// Validate record length
		if err := ValidateRecordLength(recordLength); err != nil {
			return nil, fmt.Errorf("invalid record at offset %d: %w", offset, err)
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
