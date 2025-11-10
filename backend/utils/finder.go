package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// FindByIDSequential performs a sequential scan of a binary file to find an entry by ID
// Returns the complete entry data (including ID) and nil error if found
// Returns nil and an error if not found or on file read errors
func FindByIDSequential(file *os.File, targetID uint64) ([]byte, error) {
	// Calculate header size to skip it
	headerSize := (HeaderFieldSize * 3) + 3 // 15 bytes

	// Seek to the start of the first entry (after header)
	_, err := file.Seek(int64(headerSize), 0)
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

	// Split by record separator to get individual entries
	recordSeparatorByte := []byte(RecordSeparator)[0]
	entries := bytes.Split(fileData, []byte{recordSeparatorByte})

	// Scan through each entry
	for _, entryData := range entries {
		// Skip empty entries (e.g., trailing separator creates empty slice)
		if len(entryData) == 0 {
			continue
		}

		// Each entry starts with an ID (IDSize bytes)
		if len(entryData) < IDSize {
			// Malformed entry, skip it
			continue
		}

		// Read the ID from the entry
		entryID, _, err := ReadFixedNumber(IDSize, entryData, 0)
		if err != nil {
			// Malformed entry, skip it
			continue
		}

		// Check if this is the entry we're looking for
		if entryID == targetID {
			// Found it! Return the complete entry data (including ID)
			return entryData, nil
		}
	}

	// Not found
	return nil, fmt.Errorf("entry with ID %d not found", targetID)
}
