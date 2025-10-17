package persistence

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	HeaderSize      = 4
	RecordSeparator = 0x1E
	UnitSeparator   = 0x1F
)

// WriteHeader writes the record count and separator to the file header.
func WriteHeader(writer *bufio.Writer, count uint32) error {
	if err := binary.Write(writer, binary.LittleEndian, count); err != nil {
		return fmt.Errorf("failed to write header count: %w", err)
	}

	if err := writer.WriteByte(RecordSeparator); err != nil {
		return fmt.Errorf("failed to write header separator: %w", err)
	}

	return nil
}

// ReadHeader reads and returns the record count from the file header.
func ReadHeader(reader *bufio.Reader) (uint32, error) {
	var count uint32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return 0, fmt.Errorf("failed to read header count: %w", err)
	}

	sep, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("failed to read header separator: %w", err)
	}
	if sep != RecordSeparator {
		return 0, fmt.Errorf("invalid header separator: expected 0x%02X, got 0x%02X", RecordSeparator, sep)
	}

	return count, nil
}

func InitFile(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("[DEBUG] File already exists: %s\n", filename)
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("[DEBUG] Created file: %s\n", filename)

	writer := bufio.NewWriter(file)

	// Write count = 0
	if err := binary.Write(writer, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}

	// Write record separator
	if err := writer.WriteByte(RecordSeparator); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	fmt.Printf("[DEBUG] Initialized header: [00 00 00 00 1E] (count=0)\n")
	return nil
}
