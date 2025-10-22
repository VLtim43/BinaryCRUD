package utils

import (
	"fmt"
	"os"
)

// AppendRecord appends pre-built record bytes to the binary file
// recordBytes should contain the complete record data (fields + separators)
// This function will add the final record separator and update the header
func AppendRecord(filePath string, recordBytes []byte) error {
	// Initialize the binary file if it doesn't exist
	if err := InitializeBinaryFile(filePath); err != nil {
		return fmt.Errorf("failed to initialize binary file: %w", err)
	}

	// Open file in read-write mode
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to end of file to append the new record
	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	// Write the record bytes
	if _, err := file.Write(recordBytes); err != nil {
		return fmt.Errorf("failed to write record data: %w", err)
	}

	// Write record separator to mark end of record
	if _, err := file.Write([]byte{RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	// Increment entry count in header
	if err := IncrementEntryCount(filePath); err != nil {
		return fmt.Errorf("failed to increment entry count: %w", err)
	}

	// Debug print the appended record
	DebugPrint("Wrote", string(recordBytes))

	return nil
}
