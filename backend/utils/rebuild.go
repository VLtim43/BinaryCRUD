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

// IterateEntries reads all entries from a binary file and calls the callback for each.
// Returns early if the callback returns an error.
func IterateEntries(binFilePath string, callback func(entry EntryWithOffset) error) error {
	// Check if bin file exists
	if _, err := os.Stat(binFilePath); os.IsNotExist(err) {
		return nil // No data file, nothing to iterate
	}

	// Get header size for this file
	headerSize, err := GetHeaderSize(binFilePath)
	if err != nil {
		return fmt.Errorf("failed to get header size: %w", err)
	}

	entries, err := SplitFileIntoEntries(binFilePath)
	if err != nil {
		return fmt.Errorf("failed to read entries: %w", err)
	}

	// Calculate file offsets (SplitFileIntoEntries returns position after length prefix,
	// but we need position at the length prefix for indexing)
	fileOffset := int64(headerSize)
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

// IDExtractor is a function that extracts an ID and tombstone from entry data.
// Returns (id, tombstone, error).
type IDExtractor func(data []byte) (uint64, byte, error)

// rebuildBTreeIndexGeneric is the common implementation for B+ tree index rebuilding.
func rebuildBTreeIndexGeneric(binFilePath, indexPath string, extractor IDExtractor) (*index.BTree, error) {
	tree := index.NewBTree(DefaultBTreeOrder)

	err := IterateEntries(binFilePath, func(entry EntryWithOffset) error {
		id, tombstone, err := extractor(entry.Data)
		if err == nil && tombstone == 0x00 {
			tree.Insert(id, entry.Offset)
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

// RebuildBTreeIndex scans a .bin file and rebuilds the B+ tree index for items
func RebuildBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	return rebuildBTreeIndexGeneric(binFilePath, indexPath, func(data []byte) (uint64, byte, error) {
		item, err := ParseItemEntry(data)
		if err != nil {
			return 0, 0, err
		}
		return item.ID, item.Tombstone, nil
	})
}

// RebuildCollectionBTreeIndex scans a collection .bin file and rebuilds the B+ tree index
// Works for orders.bin and promotions.bin
func RebuildCollectionBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	return rebuildBTreeIndexGeneric(binFilePath, indexPath, func(data []byte) (uint64, byte, error) {
		collection, err := ParseCollectionEntry(data)
		if err != nil {
			return 0, 0, err
		}
		return collection.ID, collection.Tombstone, nil
	})
}

// RebuildExtensibleHashIndex scans an order_promotions.bin file and rebuilds the hash index
func RebuildExtensibleHashIndex(binFilePath string, indexPath string, bucketSize int) (*index.ExtensibleHash, error) {
	hashIndex := index.NewExtensibleHash(bucketSize)

	err := IterateEntries(binFilePath, func(entry EntryWithOffset) error {
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
