package utils

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WriteFixed writes a fixed-size value with a unit separator
// size: number of bytes to write (1, 2, 4, or 8)
// value: the value to write (will be converted to uint of appropriate size)
// Example: WriteFixed(file, 2, 27) writes [00 1B][1F] (27 as 2-byte little-endian + separator)
func WriteFixed(file *os.File, size int, value uint64) error {
	// Write the value based on size
	switch size {
	case 1:
		if err := binary.Write(file, binary.LittleEndian, uint8(value)); err != nil {
			return fmt.Errorf("failed to write 1-byte value: %w", err)
		}
	case 2:
		if err := binary.Write(file, binary.LittleEndian, uint16(value)); err != nil {
			return fmt.Errorf("failed to write 2-byte value: %w", err)
		}
	case 4:
		if err := binary.Write(file, binary.LittleEndian, uint32(value)); err != nil {
			return fmt.Errorf("failed to write 4-byte value: %w", err)
		}
	case 8:
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write 8-byte value: %w", err)
		}
	default:
		return fmt.Errorf("invalid size: %d (must be 1, 2, 4, or 8)", size)
	}

	// Write unit separator
	if _, err := file.Write([]byte{UnitSeparator}); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}

	return nil
}

// WriteVariable writes a variable-length string with its size prefix and unit separator
// Format: [StringLength(4 bytes)][StringContent][UnitSeparator]
// Example: WriteVariable(file, "banana") writes [06 00 00 00][62 61 6E 61 6E 61][1F]
func WriteVariable(file *os.File, value string) error {
	// Get string bytes
	data := []byte(value)
	length := uint32(len(data))

	// Write length (4 bytes, little-endian)
	if err := binary.Write(file, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("failed to write string length: %w", err)
	}

	// Write string content
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write string content: %w", err)
	}

	// Write unit separator
	if _, err := file.Write([]byte{UnitSeparator}); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}

	return nil
}

// WriteRecordSeparator writes a record separator to mark the end of a record
func WriteRecordSeparator(file *os.File) error {
	if _, err := file.Write([]byte{RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}
	return nil
}

// RecordWriter is a function type that writes record data to a file
// The function should write all record fields but NOT the final RecordSeparator
type RecordWriter func(file *os.File) error

// AppendRecord is a generic function that handles the full write flow:
// 1. Initialize file if needed
// 2. Open file and seek to end
// 3. Call the custom writer function to write record data
// 4. Write record separator
// 5. Increment entry count in header
//
// Example usage:
//   utils.AppendRecord("data/items.bin", func(file *os.File) error {
//       if err := utils.WriteVariable(file, "banana"); err != nil {
//           return err
//       }
//       return nil
//   })
func AppendRecord(filePath string, writer RecordWriter) error {
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

	// Call the custom writer function to write record data
	if err := writer(file); err != nil {
		return fmt.Errorf("failed to write record data: %w", err)
	}

	// Write record separator to mark end of record
	if err := WriteRecordSeparator(file); err != nil {
		return err
	}

	// Increment entry count in header
	if err := IncrementEntryCount(filePath); err != nil {
		return fmt.Errorf("failed to increment entry count: %w", err)
	}

	return nil
}
