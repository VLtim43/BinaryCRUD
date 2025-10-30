package utils

import (
	"encoding/binary"
	"fmt"
	"os"
)

// ReadVariable reads a variable-length string from the file
// Format: [Length(2)][UnitSeparator][Content][UnitSeparator]
func ReadVariable(file *os.File) (string, error) {
	// Read length (2 bytes, little-endian)
	var length uint16
	if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	// Read and verify unit separator after length
	sep := make([]byte, 1)
	if _, err := file.Read(sep); err != nil {
		return "", fmt.Errorf("failed to read unit separator after length: %w", err)
	}
	if sep[0] != UnitSeparator {
		return "", fmt.Errorf("invalid unit separator after length: expected 0x1F, got 0x%X", sep[0])
	}

	// Read string content
	data := make([]byte, length)
	if _, err := file.Read(data); err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}

	// Read and verify unit separator after content
	if _, err := file.Read(sep); err != nil {
		return "", fmt.Errorf("failed to read unit separator after content: %w", err)
	}
	if sep[0] != UnitSeparator {
		return "", fmt.Errorf("invalid unit separator after content: expected 0x1F, got 0x%X", sep[0])
	}

	result := string(data)

	// Debug print the read content
	DebugPrint("Read", result)

	return result, nil
}

// SequentialRead reads all entries from a binary file sequentially
// Returns a map of record ID to record bytes (the raw bytes between record separators)
// This is file-structure agnostic - it just reads records separated by RecordSeparator
func SequentialRead(filePath string) (map[uint32][]byte, error) {
	records := make(map[uint32][]byte)

	// Open the binary file
	file, err := OpenBinaryFile(filePath)
	if err != nil {
		return records, err
	}
	defer file.Close()

	// Read header to skip past it
	_, err = ReadHeader(file)
	if err != nil {
		return records, fmt.Errorf("failed to read header: %w", err)
	}

	// Read records until EOF
	recordNum := 0
	for {
		// Read all bytes until we hit a record separator or EOF
		var recordBytes []byte
		buf := make([]byte, 1)

		for {
			n, err := file.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					// If we hit EOF and have no record bytes, we're done
					if len(recordBytes) == 0 {
						return records, nil
					}
					// If we have record bytes but hit EOF, that's an error (missing separator)
					return records, fmt.Errorf("unexpected EOF at record %d: missing record separator", recordNum)
				}
				return records, fmt.Errorf("failed to read byte at record %d: %w", recordNum, err)
			}

			if n == 0 {
				// No bytes read, we're done
				if len(recordBytes) == 0 {
					return records, nil
				}
				return records, fmt.Errorf("unexpected end of file at record %d: missing record separator", recordNum)
			}

			// Check if we hit the record separator
			if buf[0] == RecordSeparator {
				break
			}

			recordBytes = append(recordBytes, buf[0])
		}

		// If we got an empty record (consecutive separators), that's an error
		if len(recordBytes) == 0 {
			return records, fmt.Errorf("empty record at position %d", recordNum)
		}

		// Extract the actual ID from the first 4 bytes of the record
		// Format: [ID(4)][UnitSeparator][...]
		if len(recordBytes) < 4 {
			return records, fmt.Errorf("invalid record at position %d: too short to contain ID", recordNum)
		}

		actualID := binary.LittleEndian.Uint32(recordBytes[0:4])
		records[actualID] = recordBytes
		recordNum++
	}
}
