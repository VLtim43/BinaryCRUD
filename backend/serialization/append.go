package serialization

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

func AppendEntry(filename string, name string) error {
	if err := InitFile(filename); err != nil {
		return err
	}

	fmt.Printf("\n[DEBUG] === Appending Entry: \"%s\" ===\n", name)

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read current record count from header
	var count uint32
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return err
	}

	fmt.Printf("[DEBUG] Current record count: %d\n", count)

	// Seek to end of file for appending
	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	// Create the item record
	item := Item{
		Name:      name,
		Tombstone: false,
	}

	// Write the record using centralized record writer
	writer := bufio.NewWriter(file)
	if err := WriteRecord(writer, item, true); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	// Seek back to start to update header
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Update record count in header using centralized header writer
	writer = bufio.NewWriter(file)
	count++
	if err := WriteHeader(writer, count); err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	fmt.Printf("[DEBUG] Updated header count to: %d\n", count)
	fmt.Printf("[DEBUG] === Entry successfully written ===\n\n")

	return nil
}
