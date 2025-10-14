package serialization

import (
	"bufio"
	"encoding/binary"
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

	// Read record count from header
	var count uint32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return nil, err
	}

	// Skip header record separator
	if _, err := reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read record separator: %w", err)
	}

	items := make([]Item, 0, count)

	for i := uint32(0); i < count; i++ {
		// Read tombstone
		var tombstone uint8
		if err := binary.Read(reader, binary.LittleEndian, &tombstone); err != nil {
			return nil, fmt.Errorf("failed to read tombstone: %w", err)
		}

		// Skip unit separator
		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read unit separator: %w", err)
		}

		// Read name length
		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return nil, fmt.Errorf("failed to read size: %w", err)
		}

		// Skip unit separator
		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read unit separator: %w", err)
		}

		// Read name data
		data := make([]byte, size)
		if _, err := reader.Read(data); err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}

		// Skip record separator
		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read record separator: %w", err)
		}

		// Only include active records (not tombstoned)
		if tombstone == 0 {
			items = append(items, Item{
				Name:      string(data),
				Tombstone: false,
			})
		}
	}

	return items, nil
}
