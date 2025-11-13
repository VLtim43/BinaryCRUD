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
	// Lock to prevent concurrent modifications
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header to get current tombstone count
	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Split file into entries
	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Find the entry with matching ID
	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < IDSize {
			continue
		}

		// Read the ID
		entryID, _, err := ReadFixedNumber(IDSize, entryData, 0)
		if err != nil {
			continue
		}

		if entryID == id {
			// Check if already deleted
			if len(entryData) < IDSize+TombstoneSize {
				return fmt.Errorf("malformed entry for ID %d", id)
			}

			tombstone := entryData[IDSize]
			if tombstone != 0x00 {
				return fmt.Errorf("entry with ID %d is already deleted", id)
			}

			// Calculate position of tombstone byte in file
			tombstonePos := entry.Position + int64(IDSize)

			// Seek to tombstone position
			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			// Write 0x01 to mark as deleted
			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			// Force write tombstone to disk before updating header
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			// Update header to increment tombstone count (syncs internally)
			err = UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			// If an index delete function is provided, remove from index
			if indexDeleteFunc != nil {
				err = indexDeleteFunc(id)
				if err != nil {
					// Not fatal - log but continue
					// In production, this should use a logger
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
func SoftDeleteByCompositeKey(filePath string, key1, key2 uint64, mu *sync.Mutex) error {
	// Lock to prevent concurrent modifications
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	// Open file for read/write
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header to get current tombstone count
	entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Split file into entries
	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Find the entry with matching composite key
	// Format: [key1(2)][0x1F][key2(2)][0x1F][tombstone(1)]
	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < IDSize*2+TombstoneSize+2 {
			continue
		}

		offset := 0

		// Read key1
		entryKey1, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read key2
		entryKey2, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		if entryKey1 == key1 && entryKey2 == key2 {
			// Check if already deleted
			tombstone := entryData[offset]
			if tombstone != 0x00 {
				return fmt.Errorf("entry with composite key (%d, %d) is already deleted", key1, key2)
			}

			// Calculate position of tombstone byte in file
			// Position is: entry.Position + key1(2) + separator(1) + key2(2) + separator(1)
			tombstonePos := entry.Position + int64(IDSize) + 1 + int64(IDSize) + 1

			// Seek to tombstone position
			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			// Write 0x01 to mark as deleted
			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			// Force write tombstone to disk before updating header
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			// Update header to increment tombstone count (syncs internally)
			err = UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("entry with composite key (%d, %d) not found", key1, key2)
}
