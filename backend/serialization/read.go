package serialization

import (
	"bufio"
	"fmt"
	"os"
)

func ReadAllEntries(filename string) ([]Item, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []Item{}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read header using centralized header reader
	count, err := ReadHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	items := make([]Item, 0, count)

	// Read all records using centralized record reader
	for i := uint32(0); i < count; i++ {
		item, err := ReadRecord(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read record %d: %w", i+1, err)
		}

		// Only include active records (not tombstoned)
		if !item.Tombstone {
			items = append(items, *item)
		}
	}

	return items, nil
}
