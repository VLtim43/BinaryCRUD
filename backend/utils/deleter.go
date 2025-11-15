package utils

import (
	"fmt"
	"os"
	"sync"
)

// SoftDeleteByID performs a logical deletion by marking the tombstone byte as 0x01
// This function handles the common deletion pattern across all DAOs
// The optional indexDeleteFunc can be provided to also remove the entry from an index
func SoftDeleteByID(filePath string, id uint64, mu *sync.Mutex, indexDeleteFunc func(uint64) error) error {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < IDSize {
			continue
		}

		entryID, _, err := ReadFixedNumber(IDSize, entryData, 0)
		if err != nil {
			continue
		}

		if entryID == id {
			if len(entryData) < IDSize+TombstoneSize {
				return fmt.Errorf("malformed entry for ID %d", id)
			}

			tombstone := entryData[IDSize]
			if tombstone != 0x00 {
				return fmt.Errorf("entry with ID %d is already deleted", id)
			}

			tombstonePos := entry.Position + int64(IDSize)

			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			err = UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			if indexDeleteFunc != nil {
				err = indexDeleteFunc(id)
				if err != nil {
					fmt.Printf("Warning: failed to delete ID %d from index: %v\n", id, err)
				}
			}

			return nil
		}
	}

	return fmt.Errorf("entry with ID %d not found", id)
}

// SoftDeleteByCompositeKey performs a logical deletion for entries with composite keys
// Used for junction tables like order_promotions where the key is (orderID, promotionID)
// Format: [orderID(2)][promotionID(2)][tombstone(1)]
func SoftDeleteByCompositeKey(filePath string, key1, key2 uint64, mu *sync.Mutex) error {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < IDSize*2+TombstoneSize {
			continue
		}

		offset := 0

		entryKey1, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		entryKey2, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		if entryKey1 == key1 && entryKey2 == key2 {
			tombstone := entryData[offset]
			if tombstone != 0x00 {
				return fmt.Errorf("entry with composite key (%d, %d) is already deleted", key1, key2)
			}

			tombstonePos := entry.Position + int64(IDSize*2)

			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			err = UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("entry with composite key (%d, %d) not found", key1, key2)
}
