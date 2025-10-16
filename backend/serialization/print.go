package serialization

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

// PrintBinaryFile reads and prints the entire binary file in a human-readable format
// This is like "printing to paper" - showing all the binary structure as formatted text
func PrintBinaryFile(filename string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var output strings.Builder
	format := GetFormat()

	// Print file header
	output.WriteString(fmt.Sprintf("filename: %s\n", filename))

	reader := bufio.NewReader(file)

	// Read header
	offset := 0
	count, countBytes, err := readHeaderForPrint(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read header: %w", err)
	}
	output.WriteString(fmt.Sprintf("record count: %d [%s]\n", count, formatHexBytes(countBytes)))
	output.WriteString("-------------------------\n")

	offset += format.HeaderSize()

	// Read and print all records
	for i := uint32(0); i < count; i++ {
		_, bytesRead, err := printRecordSimple(&output, reader, int(i), offset)
		if err != nil {
			return "", fmt.Errorf("failed to read record %d: %w", i, err)
		}
		offset += bytesRead
	}

	outputStr := output.String()

	// Write to file
	txtFilename := filename + ".print.txt"
	if err := os.WriteFile(txtFilename, []byte(outputStr), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return outputStr, nil
}

// readHeaderForPrint reads the header and returns count and count bytes
func readHeaderForPrint(reader *bufio.Reader) (uint32, []byte, error) {
	// Read count
	var count uint32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return 0, nil, err
	}

	// Read separator (skip it)
	if _, err := reader.ReadByte(); err != nil {
		return 0, nil, err
	}

	// Format count bytes
	countBytes := []byte{
		byte(count),
		byte(count >> 8),
		byte(count >> 16),
		byte(count >> 24),
	}

	return count, countBytes, nil
}

// printRecordSimple reads and formats a single record in the simple format
func printRecordSimple(output *strings.Builder, reader *bufio.Reader, index int, offset int) (*Item, int, error) {
	format := GetFormat()
	startOffset := offset

	// Read tombstone
	var tombstone uint8
	if err := binary.Read(reader, binary.LittleEndian, &tombstone); err != nil {
		return nil, 0, err
	}
	offset += 1

	// Read unit separator (skip)
	if _, err := reader.ReadByte(); err != nil {
		return nil, 0, err
	}
	offset += 1

	// Read size
	var size uint32
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return nil, 0, err
	}
	offset += 4

	// Read unit separator (skip)
	if _, err := reader.ReadByte(); err != nil {
		return nil, 0, err
	}
	offset += 1

	// Read name data
	nameBytes := make([]byte, size)
	if _, err := reader.Read(nameBytes); err != nil {
		return nil, 0, err
	}
	offset += int(size)

	// Read unit separator (skip)
	if _, err := reader.ReadByte(); err != nil {
		return nil, 0, err
	}
	offset += 1

	// Read timestamp
	var timestamp int64
	if err := binary.Read(reader, binary.LittleEndian, &timestamp); err != nil {
		return nil, 0, err
	}
	offset += 8

	// Read record separator (skip)
	if _, err := reader.ReadByte(); err != nil {
		return nil, 0, err
	}
	offset += 1

	// Format output
	status := "active"
	if tombstone != 0 {
		status = "deleted"
	}

	tombstoneHex := fmt.Sprintf("%02X", tombstone)
	nameHex := formatHexBytes(nameBytes)
	totalSize := format.CalculateRecordSize(int(size))
	totalSizeHex := fmt.Sprintf("%02X", totalSize)

	// Format timestamp bytes and date
	timestampBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		timestampBytes[i] = byte(timestamp >> (i * 8))
	}
	timestampHex := formatHexBytes(timestampBytes)
	timestampDate := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")

	output.WriteString(fmt.Sprintf("record id: %d [%02X]\n", index, index))
	output.WriteString(fmt.Sprintf("record name: %s [%s]\n", string(nameBytes), nameHex))
	output.WriteString(fmt.Sprintf("record total size: %d [%s]\n", totalSize, totalSizeHex))
	output.WriteString(fmt.Sprintf("timestamp: %s [%s]\n", timestampDate, timestampHex))
	output.WriteString(fmt.Sprintf("status: %s [%s]\n", status, tombstoneHex))
	output.WriteString("-------------------------\n")

	item := &Item{
		Name:      string(nameBytes),
		Tombstone: tombstone != 0,
		Timestamp: timestamp,
	}

	return item, offset - startOffset, nil
}

// formatHexBytes formats a byte slice as space-separated uppercase hex values
func formatHexBytes(data []byte) string {
	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}
