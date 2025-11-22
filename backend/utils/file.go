package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateFile creates a new file and returns it open for reading and writing.
// Also creates parent directories if they don't exist.
func CreateFile(filePath string) (*os.File, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

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
// Format: [recordLength(2)][ID(2)][tombstone(1)][entry data]
func AppendEntry(file *os.File, entryWithoutId []byte) error {
	// Read current header to get nextId
	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Calculate record length (everything after the length field itself)
	// recordLength = ID(2) + tombstone(1) + len(entryWithoutId)
	recordLength := IDSize + TombstoneSize + len(entryWithoutId)

	// Generate record length field (2 bytes)
	lengthBytes, err := WriteFixedNumber(RecordLengthSize, uint64(recordLength))
	if err != nil {
		return fmt.Errorf("failed to write record length: %w", err)
	}

	// Generate ID field (2 bytes)
	idBytes, err := WriteFixedNumber(IDSize, uint64(nextId))
	if err != nil {
		return fmt.Errorf("failed to write ID: %w", err)
	}

	// Generate tombstone field (1 byte, value 0x00 for active records)
	tombstoneBytes := []byte{0x00}

	// Build the complete record: [length][ID][tombstone][entry data]
	completeRecord := make([]byte, 0, RecordLengthSize+recordLength)
	completeRecord = append(completeRecord, lengthBytes...)
	completeRecord = append(completeRecord, idBytes...)
	completeRecord = append(completeRecord, tombstoneBytes...)
	completeRecord = append(completeRecord, entryWithoutId...)

	// Seek to end of file
	_, err = file.Seek(0, 2) // 2 = io.SeekEnd
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	// Append the complete record
	err = WriteToFile(file, completeRecord)
	if err != nil {
		return fmt.Errorf("failed to write record: %w", err)
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

// AppendEntryManual appends a complete entry to the file without auto-assigning ID
// This is used for junction tables with composite keys that don't need auto-incrementing IDs
// Format: [recordLength(2)][entry data including tombstone]
// The caller is responsible for including all fields (keys, tombstone, etc.) in entryData
func AppendEntryManual(file *os.File, entryData []byte) error {
	// Read current header to get counts
	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Calculate record length (everything after the length field itself)
	recordLength := len(entryData)

	// Generate record length field (2 bytes)
	lengthBytes, err := WriteFixedNumber(RecordLengthSize, uint64(recordLength))
	if err != nil {
		return fmt.Errorf("failed to write record length: %w", err)
	}

	// Build the complete record: [length][entry data]
	completeRecord := make([]byte, 0, RecordLengthSize+recordLength)
	completeRecord = append(completeRecord, lengthBytes...)
	completeRecord = append(completeRecord, entryData...)

	// Seek to end of file
	_, err = file.Seek(0, 2) // 2 = io.SeekEnd
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	// Append the complete record
	err = WriteToFile(file, completeRecord)
	if err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	// Force write entry data to disk before updating header
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync entry to disk: %w", err)
	}

	// Update header with incremented entity count (nextId stays same for composite key tables)
	err = UpdateHeader(file, entitiesCount+1, tombstoneCount, nextId)
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
