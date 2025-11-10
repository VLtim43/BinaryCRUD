package utils

import (
	"fmt"
	"os"
)

// CreateFile creates a new file and returns it open for reading and writing.
func CreateFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("file already exists: %s", filePath)
		}
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

// WriteToFile writes binary data to the given file.
func WriteToFile(file *os.File, data []byte) error {
	_, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// WriteHeaderToFile writes a header to the beginning of a file.
// Returns an error if the file is not empty.
func WriteHeaderToFile(file *os.File, header []byte) error {
	// Check if file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() != 0 {
		return fmt.Errorf("file is not empty, cannot write header")
	}

	// Seek to the beginning to be sure
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write the header
	err = WriteToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Force write to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync header to disk: %w", err)
	}

	return nil
}

// AppendEntry appends an entry to the file with auto-assigned ID and tombstone
// Format: [ID(2)][tombstone(1)][entry data]
func AppendEntry(file *os.File, entryWithoutId []byte) error {
	// Read current header to get nextId
	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Generate ID field (2 bytes)
	idBytes, err := WriteFixedNumber(IDSize, uint64(nextId))
	if err != nil {
		return fmt.Errorf("failed to write ID: %w", err)
	}

	// Generate tombstone field (1 byte, value 0x00 for active records)
	tombstoneBytes := []byte{0x00}

	// Patch the entry with the ID and tombstone at the beginning
	patchedEntry := make([]byte, 0)
	patchedEntry = append(patchedEntry, idBytes...)
	patchedEntry = append(patchedEntry, tombstoneBytes...)
	patchedEntry = append(patchedEntry, entryWithoutId...)

	// Seek to end of file
	_, err = file.Seek(0, 2) // 2 = io.SeekEnd
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	// Append the entry
	err = WriteToFile(file, patchedEntry)
	if err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	// Append record separator
	separatorBytes, err := WriteVariable(RecordSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	err = WriteToFile(file, separatorBytes)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Force write entry data to disk before updating header
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync entry to disk: %w", err)
	}

	// Update header with incremented counts
	err = UpdateHeader(file, entitiesCount+1, tombstoneCount, nextId+1)
	if err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	// Force write header to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync header to disk: %w", err)
	}

	return nil
}
