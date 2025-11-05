package utils

import (
	"fmt"
	"os"
)

// CreateFile creates a new file and returns it open for writing.
func CreateFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("file already exists: %s", filePath)
		}
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

// WriteToFile writes a string to the given file.
func WriteToFile(file *os.File, content string) error {
	_, err := file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// WriteHeaderToFile writes a header string to the beginning of a file.
// Returns an error if the file is not empty.
func WriteHeaderToFile(file *os.File, header string) error {
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

	return nil
}
