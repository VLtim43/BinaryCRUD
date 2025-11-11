package utils

import (
	"fmt"
	"os"
)

// EntryInfo represents an entry found in the binary file
type EntryInfo struct {
	Data     []byte // The raw entry data (without record separator)
	Position int64  // File offset where this entry starts
}

// SplitFileIntoEntries reads a binary file and splits it into individual entries
// Returns a slice of EntryInfo containing the raw data and file position for each entry
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

	// Split by record separator to get individual entries
	recordSeparatorByte := []byte(RecordSeparator)[0]
	entries := make([]EntryInfo, 0)

	entryStart := HeaderSize

	for i := HeaderSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, EntryInfo{
				Data:     fileData[entryStart:i],
				Position: int64(entryStart),
			})
			entryStart = i + 1
		}
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

	// Write empty header (0 entities, 0 tombstones, nextId=0)
	header, err := WriteHeader(0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}
