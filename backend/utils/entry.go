package utils

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Check minimum size for header
	if len(fileData) < MagicSize+FilenameLengthSize {
		return []EntryInfo{}, nil
	}

	// Get actual header size by reading filename length
	filenameLen := int(fileData[MagicSize])
	headerSize := CalculateHeaderSize(string(fileData[MagicSize+FilenameLengthSize : MagicSize+FilenameLengthSize+filenameLen]))

	if len(fileData) < headerSize {
		return []EntryInfo{}, nil
	}

	entries := make([]EntryInfo, 0)
	offset := headerSize

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
// The filename is extracted from the filePath and stored in the header
func EnsureFileExists(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Extract just the filename from the path
	filename := filepath.Base(filePath)

	// Create the file
	file, err := CreateFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write empty header with filename
	header, err := WriteHeader(filename, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// GetHeaderSize reads a file and returns its header size
func GetHeaderSize(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Read magic + filename length
	header := make([]byte, MagicSize+FilenameLengthSize)
	n, err := file.Read(header)
	if err != nil || n < MagicSize+FilenameLengthSize {
		return 0, fmt.Errorf("failed to read header")
	}

	filenameLen := int(header[MagicSize])
	return CalculateHeaderSize(string(make([]byte, filenameLen))), nil
}

// GetHeaderSizeFromFile reads header size from an open file (resets position)
func GetHeaderSizeFromFile(file *os.File) (int, error) {
	// Save current position
	currentPos, err := file.Seek(0, 1)
	if err != nil {
		return 0, err
	}

	// Seek to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	// Read magic + filename length
	header := make([]byte, MagicSize+FilenameLengthSize)
	n, err := file.Read(header)
	if err != nil || n < MagicSize+FilenameLengthSize {
		file.Seek(currentPos, 0) // Restore position
		return 0, fmt.Errorf("failed to read header")
	}

	filenameLen := int(header[MagicSize])

	// Read filename to get actual size
	filenameBytes := make([]byte, filenameLen)
	n, err = file.Read(filenameBytes)
	if err != nil || n < filenameLen {
		file.Seek(currentPos, 0)
		return 0, fmt.Errorf("failed to read filename")
	}

	// Restore original position
	file.Seek(currentPos, 0)

	return CalculateHeaderSize(string(filenameBytes)), nil
}
