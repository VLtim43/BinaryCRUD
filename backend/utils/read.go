package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
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

	return string(data), nil
}

// FormatBytes converts a byte slice to a hex string representation
func FormatBytes(data []byte) string {
	parts := make([]string, len(data))
	for i, b := range data {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// PrintBinaryFile prints the complete structure of a binary file with header and records
func PrintBinaryFile(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var output strings.Builder

	// Print filename
	output.WriteString(fmt.Sprintf("Filename: %s\n\n", filePath))

	// Read and print header
	header, err := ReadHeader(file)
	if err != nil {
		return "", fmt.Errorf("failed to read header: %w", err)
	}

	// Entry Count
	entryCountBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(entryCountBytes, header.EntryCount)
	output.WriteString(fmt.Sprintf("Entries Count: %d  %s\n", header.EntryCount, FormatBytes(entryCountBytes)))

	// Tombstone Count
	tombstoneBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(tombstoneBytes, header.TombstoneCount)
	output.WriteString(fmt.Sprintf("Tombstone Count: %d  %s\n", header.TombstoneCount, FormatBytes(tombstoneBytes)))

	output.WriteString("════════════════════════════════════════════\n")

	// Read and print records
	if header.EntryCount == 0 {
		output.WriteString("(No records)\n")
	} else {
		for i := uint32(0); i < header.EntryCount; i++ {
			// Read the item name
			itemName, err := ReadVariable(file)
			if err != nil {
				output.WriteString(fmt.Sprintf("ERROR: %v\n", err))
				break
			}

			// Get the item name bytes for display
			itemBytes := []byte(itemName)

			output.WriteString(fmt.Sprintf("Item Name: \"%s\"  %s\n", itemName, FormatBytes(itemBytes)))

			// Read record separator
			sep := make([]byte, 1)
			if _, err := file.Read(sep); err != nil {
				output.WriteString(fmt.Sprintf("ERROR reading record separator: %v\n", err))
				break
			}

			output.WriteString("────────────────────────────────────────────\n")
		}
	}

	return output.String(), nil
}
