package utils

import (
	"BinaryCRUD/backend/index"
	"fmt"
	"os"
)

// EntryWithOffset contains entry data and its file offset
type EntryWithOffset struct {
	Data   []byte
	Offset int64
}

// iterateEntries reads all entries from a binary file and calls the callback for each
// Returns early if the callback returns an error
func iterateEntries(binFilePath string, callback func(entry EntryWithOffset) error) error {
	// Check if bin file exists
	if _, err := os.Stat(binFilePath); os.IsNotExist(err) {
		return nil // No data file, nothing to iterate
	}

	entries, err := SplitFileIntoEntries(binFilePath)
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	// Calculate file offsets (SplitFileIntoEntries returns position after length prefix,
	// but we need position at the length prefix for indexing)
	fileOffset := int64(HeaderSize)
	for _, entry := range entries {
		if err := callback(EntryWithOffset{
			Data:   entry.Data,
			Offset: fileOffset,
		}); err != nil {
			return err
		}
		// Move to next record: length prefix + data length
		fileOffset += int64(RecordLengthSize + len(entry.Data))
	}

	return nil
}

// RebuildBTreeIndex scans a .bin file and rebuilds the B+ tree index for items
func RebuildBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	tree := index.NewBTree(DefaultBTreeOrder)

	err := iterateEntries(binFilePath, func(entry EntryWithOffset) error {
		item, err := ParseItemEntry(entry.Data)
		if err == nil && item.Tombstone == 0x00 {
			tree.Insert(item.ID, entry.Offset)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := tree.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return tree, nil
}

// RebuildCollectionBTreeIndex scans a collection .bin file and rebuilds the B+ tree index
// Works for orders.bin and promotions.bin
func RebuildCollectionBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	tree := index.NewBTree(DefaultBTreeOrder)

	err := iterateEntries(binFilePath, func(entry EntryWithOffset) error {
		collection, err := ParseCollectionEntry(entry.Data)
		if err == nil && collection.Tombstone == 0x00 {
			tree.Insert(collection.ID, entry.Offset)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := tree.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return tree, nil
}

// RebuildExtensibleHashIndex scans an order_promotions.bin file and rebuilds the hash index
func RebuildExtensibleHashIndex(binFilePath string, indexPath string, bucketSize int) (*index.ExtensibleHash, error) {
	hashIndex := index.NewExtensibleHash(bucketSize)

	err := iterateEntries(binFilePath, func(entry EntryWithOffset) error {
		op, err := ParseOrderPromotionEntry(entry.Data)
		if err == nil && op.Tombstone == 0x00 {
			hashIndex.Insert(op.OrderID, op.PromotionID, entry.Offset)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := hashIndex.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return hashIndex, nil
}
