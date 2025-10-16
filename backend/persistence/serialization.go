package persistence

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type Item struct {
	RecordID  uint32
	Name      string
	Tombstone bool
	Timestamp int64 // Unix timestamp in seconds
}

const (
	HeaderSize      = 4
	RecordSeparator = 0x1E
	UnitSeparator   = 0x1F
)

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
