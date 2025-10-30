package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// LoggerInterface defines the interface for logging
type LoggerInterface interface {
	Log(message string)
}

var globalLogger LoggerInterface

// SetLogger sets the global logger instance
func SetLogger(logger LoggerInterface) {
	globalLogger = logger
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
// This is the user-facing print function (called when clicking the "Print Binary File" button)
func PrintBinaryFile(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	var output strings.Builder

	// Print filename
	output.WriteString(fmt.Sprintf("Filename: %s\n\n", filePath))

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header info
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

	// Next ID
	nextIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nextIDBytes, header.NextID)
	output.WriteString(fmt.Sprintf("Next ID: %d  %s\n", header.NextID, FormatBytes(nextIDBytes)))

	output.WriteString("════════════════════════════════════════════\n")

	// Scan file for records separated by RecordSeparator
	recordNum := 0
	buf := make([]byte, 1)
	recordBytes := []byte{}

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}

		if buf[0] == RecordSeparator {
			// Found end of record
			if len(recordBytes) > 0 {
				output.WriteString(fmt.Sprintf("Record #%d (%d bytes):\n", recordNum, len(recordBytes)))
				output.WriteString(FormatBytes(recordBytes) + "\n")
				output.WriteString("────────────────────────────────────────────\n")
				recordNum++
				recordBytes = []byte{}
			}
		} else {
			recordBytes = append(recordBytes, buf[0])
		}
	}

	if recordNum == 0 {
		output.WriteString("(No records)\n")
	}

	return output.String(), nil
}

// DebugPrint prints a debug message with optional formatting
// Single parameter: DebugPrint("Info message") → [DEBUGGER] Info message
// With formatting: DebugPrint("Deleted %d files from %s", count, dir) → [DEBUGGER] Deleted 5 files from data
// Two string parameters: DebugPrint("Operation", "data") → [DEBUGGER] Operation: "data" [binary]
func DebugPrint(format string, args ...interface{}) {
	var message string

	// Check if we have exactly one string argument and format doesn't contain %
	if len(args) == 1 {
		if str, ok := args[0].(string); ok && !containsFormatVerb(format) {
			// Two string parameters: operation + content (show binary)
			contentBytes := []byte(str)
			message = fmt.Sprintf("[DEBUGGER] %s: \"%s\" %s\n", format, str, FormatBytes(contentBytes))
			fmt.Print(message)
			if globalLogger != nil {
				globalLogger.Log(message)
			}
			return
		}
	}

	// Info message with optional formatting
	if len(args) > 0 {
		message = fmt.Sprintf("[DEBUGGER] "+format+"\n", args...)
	} else {
		message = fmt.Sprintf("[DEBUGGER] %s\n", format)
	}

	fmt.Print(message)
	if globalLogger != nil {
		globalLogger.Log(message)
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
