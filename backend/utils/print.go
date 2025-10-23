package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// FormatBytes converts a byte slice to a hex string representation
func FormatBytes(data []byte) string {
	parts := make([]string, len(data))
	for i, b := range data {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// PrintBinaryFile prints the complete structure of a binary file with header and records
// This is the user-facing print function (called when clicking the "Print Binary File" button)
func PrintBinaryFile(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	var output strings.Builder

	// Print filename
	output.WriteString(fmt.Sprintf("Filename: %s\n\n", filePath))

	// Read header info
	header, err := GetHeaderInfo(filePath)
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

	// Read and print records using SequentialRead
	if header.EntryCount == 0 {
		output.WriteString("(No records)\n")
	} else {
		records, err := SequentialRead(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read records: %w", err)
		}

		// Print records in order
		for i := uint32(0); i < header.EntryCount; i++ {
			recordBytes, exists := records[i]
			if !exists {
				output.WriteString(fmt.Sprintf("ERROR: Missing record at index %d\n", i))
				break
			}

			// Parse the variable-length string from the record bytes
			// Format: [Length(2)][UnitSeparator][Content][UnitSeparator]
			if len(recordBytes) < 4 {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: too short\n", i))
				break
			}

			// Extract length (first 2 bytes, little-endian)
			length := uint16(recordBytes[0]) | uint16(recordBytes[1])<<8

			// Verify first unit separator at position 2
			if recordBytes[2] != UnitSeparator {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: missing first unit separator\n", i))
				break
			}

			// Extract content
			contentStart := 3
			contentEnd := contentStart + int(length)

			if contentEnd > len(recordBytes) {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: content length mismatch\n", i))
				break
			}

			itemName := string(recordBytes[contentStart:contentEnd])
			itemBytes := []byte(itemName)

			output.WriteString(fmt.Sprintf("Item Name: \"%s\"  %s\n", itemName, FormatBytes(itemBytes)))
			output.WriteString("────────────────────────────────────────────\n")
		}
	}

	return output.String(), nil
}

// DebugPrint prints a debug message with optional formatting
// Single parameter: DebugPrint("Info message") → [DEBUGGER] Info message
// With formatting: DebugPrint("Deleted %d files from %s", count, dir) → [DEBUGGER] Deleted 5 files from data
// Two string parameters: DebugPrint("Operation", "data") → [DEBUGGER] Operation: "data" [binary]
func DebugPrint(format string, args ...interface{}) {
	// Check if we have exactly one string argument and format doesn't contain %
	if len(args) == 1 {
		if str, ok := args[0].(string); ok && !containsFormatVerb(format) {
			// Two string parameters: operation + content (show binary)
			contentBytes := []byte(str)
			fmt.Printf("[DEBUGGER] %s: \"%s\" %s\n", format, str, FormatBytes(contentBytes))
			return
		}
	}

	// Info message with optional formatting
	if len(args) > 0 {
		fmt.Printf("[DEBUGGER] "+format+"\n", args...)
	} else {
		fmt.Printf("[DEBUGGER] %s\n", format)
	}
}

// containsFormatVerb checks if a string contains printf format verbs
func containsFormatVerb(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' && s[i+1] != '%' {
			return true
		}
	}
	return false
}
