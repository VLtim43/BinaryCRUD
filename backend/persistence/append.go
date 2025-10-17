package persistence

import (
	"fmt"
	"time"
)

// AppendResult contains the result of appending an entry.
type AppendResult struct {
	RecordID uint32 // The ID of the newly created record (0-based)
	Offset   int64  // The byte offset where the record was written
}

// AppendItem appends a new item record to the backing binary file.
func AppendItem(filename string, name string) (*AppendResult, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot add empty item: name is required")
	}

	fmt.Printf("\n[DEBUG] === Appending Item: \"%s\" ===\n", name)

	result, currentCount, err := appendBinaryRecord(
		filename,
		"item record",
		true,
		func(nextID uint32) (Item, error) {
			return Item{
				RecordID:  nextID,
				Name:      name,
				Tombstone: false,
				Timestamp: time.Now().Unix(),
			}, nil
		},
		WriteItemRecord,
		func(count uint32, offset int64) {
			fmt.Printf("[DEBUG] Current record count: %d\n", count)
			fmt.Printf("[DEBUG] Writing record at offset: %d\n", offset)
		},
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] Updated header: count=%d\n", currentCount+1)
	fmt.Printf("[DEBUG] Assigned recordID: %d at offset: %d\n", result.RecordID, result.Offset)
	fmt.Printf("[DEBUG] === Item successfully written ===\n\n")

	return result, nil
}
