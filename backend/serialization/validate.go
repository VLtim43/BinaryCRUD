package serialization

import (
	"bufio"
	"fmt"
	"os"
)

// ValidationResult contains the results of file validation
type ValidationResult struct {
	Valid           bool
	Errors          []string
	Warnings        []string
	HeaderCount     uint32
	ActualRecords   uint32
	CorruptedAt     int64 // Byte offset where corruption was detected
}

// ValidateFile performs comprehensive validation of the binary file structure
func ValidateFile(filename string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check if file exists
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file does not exist: %s", filename))
		return result, nil
	}

	// Check minimum file size (must have at least header)
	format := GetFormat()
	minSize := int64(format.HeaderSize())
	if fileInfo.Size() < minSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file too small: %d bytes (minimum: %d bytes)", fileInfo.Size(), minSize))
		return result, nil
	}

	// Open file for reading
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Validate header
	count, err := ReadHeader(reader)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid header: %v", err))
		return result, nil
	}
	result.HeaderCount = count

	// Validate records
	actualRecords := uint32(0)
	offset := int64(format.HeaderSize())

	for i := uint32(0); i < count; i++ {
		item, err := ReadRecord(reader)
		if err != nil {
			result.Valid = false
			result.CorruptedAt = offset
			result.Errors = append(result.Errors, fmt.Sprintf("corrupted at record %d (offset %d): %v", i+1, offset, err))
			break
		}

		actualRecords++
		recordSize := format.CalculateRecordSize(len(item.Name))
		offset += int64(recordSize)

		// Validate record data
		if len(item.Name) == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("record %d has empty name", i+1))
		}

		// Check for very long names (potential corruption)
		if len(item.Name) > 10000 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("record %d has unusually long name: %d bytes", i+1, len(item.Name)))
		}
	}

	result.ActualRecords = actualRecords

	// Verify record count matches
	if actualRecords != count {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("record count mismatch: header says %d, found %d", count, actualRecords))
	}

	// Check for extra data after last record
	extraBytes, err := reader.Peek(1)
	if err == nil && len(extraBytes) > 0 {
		result.Warnings = append(result.Warnings, "file contains extra data after last record")
	}

	return result, nil
}

// ValidateFileStructure is a convenience function that validates and returns a boolean
func ValidateFileStructure(filename string) (bool, error) {
	result, err := ValidateFile(filename)
	if err != nil {
		return false, err
	}
	return result.Valid, nil
}

// PrintValidationResult formats the validation result as a human-readable string
func PrintValidationResult(result *ValidationResult) string {
	if result.Valid {
		return fmt.Sprintf("✓ File is valid\n  Header Count: %d\n  Actual Records: %d\n",
			result.HeaderCount, result.ActualRecords)
	}

	output := "✗ File validation failed\n"
	output += fmt.Sprintf("  Header Count: %d\n", result.HeaderCount)
	output += fmt.Sprintf("  Actual Records Found: %d\n", result.ActualRecords)

	if len(result.Errors) > 0 {
		output += "\nErrors:\n"
		for _, err := range result.Errors {
			output += fmt.Sprintf("  - %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		output += "\nWarnings:\n"
		for _, warn := range result.Warnings {
			output += fmt.Sprintf("  - %s\n", warn)
		}
	}

	if result.CorruptedAt > 0 {
		output += fmt.Sprintf("\nCorruption detected at byte offset: %d (0x%X)\n",
			result.CorruptedAt, result.CorruptedAt)
	}

	return output
}
