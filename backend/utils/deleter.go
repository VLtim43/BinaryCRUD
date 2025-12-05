package utils

import (
	"fmt"
	"os"
	"sync"
)

// entryMatcher defines how to match an entry and where its tombstone is located
type entryMatcher struct {
	match        func(entryData []byte) bool
	tombstonePos func(entryPos int64) int64
	minSize      int
	notFoundErr  string
	alreadyDelErr string
}

// softDeleteCore is the shared implementation for soft deletion
func softDeleteCore(filePath string, mu *sync.Mutex, matcher entryMatcher, onDelete func(uint64) error) error {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, entitiesCount, tombstoneCount, nextId, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < matcher.minSize {
			continue
		}

		if !matcher.match(entryData) {
			continue
		}

		// Found the entry - check tombstone
		tombstoneOffset := matcher.tombstonePos(0) // relative offset within entry
		if entryData[tombstoneOffset] != 0x00 {
			return fmt.Errorf(matcher.alreadyDelErr)
		}

		// Write tombstone
		tombstoneFilePos := entry.Position + matcher.tombstonePos(0)
		if _, err = file.Seek(tombstoneFilePos, 0); err != nil {
			return fmt.Errorf("failed to seek to tombstone: %w", err)
		}
		if _, err = file.Write([]byte{0x01}); err != nil {
			return fmt.Errorf("failed to write tombstone: %w", err)
		}
		if err = file.Sync(); err != nil {
			return fmt.Errorf("failed to sync tombstone to disk: %w", err)
		}

		if err = UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId); err != nil {
			return fmt.Errorf("failed to update header: %w", err)
		}

		if onDelete != nil {
			// Extract ID for callback (first IDSize bytes)
			id, _, _ := ReadFixedNumber(IDSize, entryData, 0)
			if err = onDelete(id); err != nil {
				fmt.Printf("Warning: onDelete callback failed: %v\n", err)
			}
		}

		return nil
	}

	return fmt.Errorf(matcher.notFoundErr)
}

// SoftDeleteByID performs a logical deletion by marking the tombstone byte as 0x01
// This function handles the common deletion pattern across all DAOs
// The optional indexDeleteFunc can be provided to also remove the entry from an index
func SoftDeleteByID(filePath string, id uint64, mu *sync.Mutex, indexDeleteFunc func(uint64) error) error {
	matcher := entryMatcher{
		match: func(entryData []byte) bool {
			entryID, _, err := ReadFixedNumber(IDSize, entryData, 0)
			return err == nil && entryID == id
		},
		tombstonePos:  func(_ int64) int64 { return int64(IDSize) },
		minSize:       IDSize + TombstoneSize,
		notFoundErr:   fmt.Sprintf("entry with ID %d not found", id),
		alreadyDelErr: fmt.Sprintf("entry with ID %d is already deleted", id),
	}
	return softDeleteCore(filePath, mu, matcher, indexDeleteFunc)
}

// SoftDeleteByCompositeKey performs a logical deletion for entries with composite keys
// Used for junction tables like order_promotions where the key is (orderID, promotionID)
// Format: [orderID(2)][promotionID(2)][tombstone(1)]
func SoftDeleteByCompositeKey(filePath string, key1, key2 uint64, mu *sync.Mutex) error {
	matcher := entryMatcher{
		match: func(entryData []byte) bool {
			entryKey1, offset, err := ReadFixedNumber(IDSize, entryData, 0)
			if err != nil {
				return false
			}
			entryKey2, _, err := ReadFixedNumber(IDSize, entryData, offset)
			if err != nil {
				return false
			}
			return entryKey1 == key1 && entryKey2 == key2
		},
		tombstonePos:  func(_ int64) int64 { return int64(IDSize * 2) },
		minSize:       IDSize*2 + TombstoneSize,
		notFoundErr:   fmt.Sprintf("entry with composite key (%d, %d) not found", key1, key2),
		alreadyDelErr: fmt.Sprintf("entry with composite key (%d, %d) is already deleted", key1, key2),
	}
	return softDeleteCore(filePath, mu, matcher, nil)
}
