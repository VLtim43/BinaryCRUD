package serialization

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// WriteHeader writes the file header with the given record count
func WriteHeader(writer *bufio.Writer, count uint32) error {
	// Write record count (4 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, count); err != nil {
		return fmt.Errorf("failed to write header count: %w", err)
	}

	// Write record separator
	if err := writer.WriteByte(RecordSeparator); err != nil {
		return fmt.Errorf("failed to write header separator: %w", err)
	}

	return nil
}

// ReadHeader reads and returns the record count from the file header
func ReadHeader(reader *bufio.Reader) (count uint32, err error) {
	// Read record count (4 bytes, little-endian)
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return 0, fmt.Errorf("failed to read header count: %w", err)
	}

	// Read and verify record separator
	sep, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("failed to read header separator: %w", err)
	}
	if sep != RecordSeparator {
		return 0, fmt.Errorf("invalid header separator: expected 0x%02X, got 0x%02X", RecordSeparator, sep)
	}

	return count, nil
}

// WriteRecord writes a single record to the writer
// This is the single source of truth for record format during writes
func WriteRecord(writer *bufio.Writer, item Item, debug bool) error {
	nameBytes := []byte(item.Name)
	size := uint32(len(nameBytes))

	// Write ID (4 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, item.RecordID); err != nil {
		return fmt.Errorf("failed to write record ID: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote record ID: [%02X %02X %02X %02X] (%d)\n",
			byte(item.RecordID), byte(item.RecordID>>8), byte(item.RecordID>>16), byte(item.RecordID>>24), item.RecordID)
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write tombstone (0 = active, 1 = deleted)
	tombstone := uint8(0)
	if item.Tombstone {
		tombstone = 1
	}
	if err := binary.Write(writer, binary.LittleEndian, tombstone); err != nil {
		return fmt.Errorf("failed to write tombstone: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote tombstone: [%02X] (%s)\n", tombstone, getTombstoneStatus(item.Tombstone))
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write name length (4 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, size); err != nil {
		return fmt.Errorf("failed to write name size: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote name length: [%02X %02X %02X %02X] (%d bytes)\n",
			byte(size), byte(size>>8), byte(size>>16), byte(size>>24), size)
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write name data
	if _, err := writer.Write(nameBytes); err != nil {
		return fmt.Errorf("failed to write name data: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote name data: [%s] (\"%s\")\n", formatHexBytesDebug(nameBytes), item.Name)
	}

	// Write unit separator before timestamp
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write timestamp (8 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, item.Timestamp); err != nil {
		return fmt.Errorf("failed to write timestamp: %w", err)
	}
	if debug {
		timestampBytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			timestampBytes[i] = byte(item.Timestamp >> (i * 8))
		}
		timestampDate := time.Unix(item.Timestamp, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("[DEBUG] Wrote timestamp: [%s] (%s)\n", formatHexBytesDebug(timestampBytes), timestampDate)
	}

	// Write record separator
	if err := writer.WriteByte(RecordSeparator); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote record separator: [%02X]\n", RecordSeparator)
	}

	return nil
}

// ReadRecord reads a single record from the reader
// This is the single source of truth for record format during reads
func ReadRecord(reader *bufio.Reader) (*Item, error) {
	// Read ID
	var recordID uint32
	if err := binary.Read(reader, binary.LittleEndian, &recordID); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read record ID: %w", err)
	}

	// Read and verify unit separator
	sep, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after ID: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after ID: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read tombstone
	var tombstone uint8
	if err := binary.Read(reader, binary.LittleEndian, &tombstone); err != nil {
		return nil, fmt.Errorf("failed to read tombstone: %w", err)
	}

	// Read and verify unit separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after tombstone: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after tombstone: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read name size (4 bytes, little-endian)
	var size uint32
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return nil, fmt.Errorf("failed to read name size: %w", err)
	}

	// Sanity check: name size shouldn't be larger than 1MB (prevents corruption issues)
	if size > 1024*1024 {
		return nil, fmt.Errorf("invalid name size: %d (possibly corrupted file)", size)
	}

	// Read and verify unit separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after size: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after size: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read name data
	nameBytes := make([]byte, size)
	if _, err := io.ReadFull(reader, nameBytes); err != nil {
		return nil, fmt.Errorf("failed to read name data: %w", err)
	}

	// Read and verify unit separator before timestamp
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator before timestamp: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator before timestamp: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read timestamp (8 bytes, little-endian)
	var timestamp int64
	if err := binary.Read(reader, binary.LittleEndian, &timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}

	// Read and verify record separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read record separator: %w", err)
	}
	if sep != RecordSeparator {
		return nil, fmt.Errorf("invalid record separator: expected 0x%02X, got 0x%02X", RecordSeparator, sep)
	}

	return &Item{
		RecordID:  recordID,
		Name:      string(nameBytes),
		Tombstone: tombstone != 0,
		Timestamp: timestamp,
	}, nil
}

func getTombstoneStatus(tombstone bool) string {
	if tombstone {
		return "deleted"
	}
	return "active"
}

// formatHexBytesDebug formats a byte slice as space-separated uppercase hex values
func formatHexBytesDebug(data []byte) string {
	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}
