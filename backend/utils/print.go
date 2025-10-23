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

	// Next ID
	nextIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nextIDBytes, header.NextID)
	output.WriteString(fmt.Sprintf("Next ID: %d  %s\n", header.NextID, FormatBytes(nextIDBytes)))

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

			// Parse the record bytes
			// Format: [ID(4)][UnitSeparator][StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
			if len(recordBytes) < 8 { // Minimum: 4 bytes ID + 1 sep + 2 bytes length + 1 sep
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: too short\n", i))
				break
			}

			// Extract ID (first 4 bytes, little-endian)
			itemID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24
			idBytes := recordBytes[0:4]

			// Verify unit separator after ID
			if recordBytes[4] != UnitSeparator {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: missing separator after ID\n", i))
				break
			}

			// Extract length (bytes 5-6, little-endian)
			length := uint16(recordBytes[5]) | uint16(recordBytes[6])<<8

			// Verify unit separator after length
			if recordBytes[7] != UnitSeparator {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: missing separator after length\n", i))
				break
			}

			// Extract content
			contentStart := 8
			contentEnd := contentStart + int(length)

			if contentEnd > len(recordBytes) {
				output.WriteString(fmt.Sprintf("ERROR: Invalid record at index %d: content length mismatch\n", i))
				break
			}

			itemName := string(recordBytes[contentStart:contentEnd])
			itemBytes := []byte(itemName)

			output.WriteString(fmt.Sprintf("Item ID: %d  %s\n", itemID, FormatBytes(idBytes)))
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
