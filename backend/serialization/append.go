package serialization

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// AppendResult contains the result of appending an entry
type AppendResult struct {
	RecordID uint32 // The ID of the newly created record (0-based)
	Offset   int64  // The byte offset where the record was written
}

func AppendEntry(filename string, name string) (*AppendResult, error) {
	// Validate that name is not empty
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot add empty item: name is required")
	}

	if err := InitFile(filename); err != nil {
		return nil, err
	}

	fmt.Printf("\n[DEBUG] === Appending Entry: \"%s\" ===\n", name)

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read current record count from header
	reader := bufio.NewReader(file)
	count, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] Current record count: %d\n", count)

	// Seek to end of file for appending and capture offset
	offset, err := file.Seek(0, 2)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] Writing record at offset: %d\n", offset)

	// Create the item record with current timestamp
	// RecordID is assigned from the current count (before incrementing)
	item := Item{
		RecordID:  count,
		Name:      name,
		Tombstone: false,
		Timestamp: time.Now().Unix(),
	}

	// Write the record using centralized record writer
	writer := bufio.NewWriter(file)
	if err := WriteRecord(writer, item, true); err != nil {
		return nil, fmt.Errorf("failed to write record: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	// Seek back to start to update header
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	// Update record count in header using centralized header writer
	recordID := count
	count++

	writer = bufio.NewWriter(file)
	if err := WriteHeader(writer, count); err != nil {
		return nil, fmt.Errorf("failed to update header: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] Updated header: count=%d\n", count)
	fmt.Printf("[DEBUG] Assigned recordID: %d at offset: %d\n", recordID, offset)
	fmt.Printf("[DEBUG] === Entry successfully written ===\n\n")

	return &AppendResult{
		RecordID: recordID,
		Offset:   offset,
	}, nil
}
