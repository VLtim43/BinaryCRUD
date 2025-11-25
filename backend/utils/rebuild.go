package utils

import (
	"BinaryCRUD/backend/index"
	"fmt"
	"io"
	"os"
)

// RebuildBTreeIndex scans a .bin file and rebuilds the B+ tree index
// This is used for crash recovery when the index file is missing or corrupted
func RebuildBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	tree := index.NewBTree(DefaultBTreeOrder)

	// Check if bin file exists
	if _, err := os.Stat(binFilePath); os.IsNotExist(err) {
		// No data file, return empty tree
		return tree, nil
	}

	// Open the bin file
	file, err := os.Open(binFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bin file: %w", err)
	}
	defer file.Close()

	// Read header to verify file is valid
	_, _, _, err = ReadHeader(file)
	if err != nil {
		// Corrupted or empty file, return empty tree
		return tree, nil
	}

	// Seek past header
	_, err = file.Seek(int64(HeaderSize), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past header: %w", err)
	}

	// Read all file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Track current file offset (relative to after header)
	fileOffset := int64(HeaderSize)
	dataOffset := 0

	// Parse all records and rebuild index
	for dataOffset < len(fileData) {
		// Check if we have enough bytes for the length field
		if dataOffset+RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, lengthEnd, err := ReadFixedNumber(RecordLengthSize, fileData, dataOffset)
		if err != nil {
			break
		}

		// Check if we have enough bytes for the complete record
		if lengthEnd+int(recordLength) > len(fileData) {
			break
		}

		// Extract the record data (without length prefix)
		entryData := fileData[lengthEnd : lengthEnd+int(recordLength)]

		// Parse to get ID and tombstone
		item, err := ParseItemEntry(entryData)
		if err == nil && item.Tombstone == 0x00 {
			// Only index non-deleted entries
			tree.Insert(item.ID, fileOffset)
		}

		// Move to next record
		recordTotalSize := RecordLengthSize + int(recordLength)
		fileOffset += int64(recordTotalSize)
		dataOffset = lengthEnd + int(recordLength)
	}

	// Save the rebuilt index
	if err := tree.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return tree, nil
}

// RebuildCollectionBTreeIndex scans a collection .bin file and rebuilds the B+ tree index
// Works for orders.bin and promotions.bin
func RebuildCollectionBTreeIndex(binFilePath string, indexPath string) (*index.BTree, error) {
	tree := index.NewBTree(DefaultBTreeOrder)

	// Check if bin file exists
	if _, err := os.Stat(binFilePath); os.IsNotExist(err) {
		// No data file, return empty tree
		return tree, nil
	}

	// Open the bin file
	file, err := os.Open(binFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bin file: %w", err)
	}
	defer file.Close()

	// Read header to verify file is valid
	_, _, _, err = ReadHeader(file)
	if err != nil {
		// Corrupted or empty file, return empty tree
		return tree, nil
	}

	// Seek past header
	_, err = file.Seek(int64(HeaderSize), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past header: %w", err)
	}

	// Read all file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Track current file offset (relative to after header)
	fileOffset := int64(HeaderSize)
	dataOffset := 0

	// Parse all records and rebuild index
	for dataOffset < len(fileData) {
		// Check if we have enough bytes for the length field
		if dataOffset+RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, lengthEnd, err := ReadFixedNumber(RecordLengthSize, fileData, dataOffset)
		if err != nil {
			break
		}

		// Check if we have enough bytes for the complete record
		if lengthEnd+int(recordLength) > len(fileData) {
			break
		}

		// Extract the record data (without length prefix)
		entryData := fileData[lengthEnd : lengthEnd+int(recordLength)]

		// Parse to get ID and tombstone
		collection, err := ParseCollectionEntry(entryData)
		if err == nil && collection.Tombstone == 0x00 {
			// Only index non-deleted entries
			tree.Insert(collection.ID, fileOffset)
		}

		// Move to next record
		recordTotalSize := RecordLengthSize + int(recordLength)
		fileOffset += int64(recordTotalSize)
		dataOffset = lengthEnd + int(recordLength)
	}

	// Save the rebuilt index
	if err := tree.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return tree, nil
}

// RebuildExtensibleHashIndex scans an order_promotions.bin file and rebuilds the hash index
func RebuildExtensibleHashIndex(binFilePath string, indexPath string, bucketSize int) (*index.ExtensibleHash, error) {
	hashIndex := index.NewExtensibleHash(bucketSize)

	// Check if bin file exists
	if _, err := os.Stat(binFilePath); os.IsNotExist(err) {
		// No data file, return empty hash
		return hashIndex, nil
	}

	// Open the bin file
	file, err := os.Open(binFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bin file: %w", err)
	}
	defer file.Close()

	// Read header to verify file is valid
	_, _, _, err = ReadHeader(file)
	if err != nil {
		// Corrupted or empty file, return empty hash
		return hashIndex, nil
	}

	// Seek past header
	_, err = file.Seek(int64(HeaderSize), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past header: %w", err)
	}

	// Read all file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Track current file offset (relative to after header)
	fileOffset := int64(HeaderSize)
	dataOffset := 0

	// Parse all records and rebuild index
	for dataOffset < len(fileData) {
		// Check if we have enough bytes for the length field
		if dataOffset+RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, lengthEnd, err := ReadFixedNumber(RecordLengthSize, fileData, dataOffset)
		if err != nil {
			break
		}

		// Check if we have enough bytes for the complete record
		if lengthEnd+int(recordLength) > len(fileData) {
			break
		}

		// Extract the record data (without length prefix)
		entryData := fileData[lengthEnd : lengthEnd+int(recordLength)]

		// Parse to get IDs and tombstone
		op, err := ParseOrderPromotionEntry(entryData)
		if err == nil && op.Tombstone == 0x00 {
			// Only index non-deleted entries
			hashIndex.Insert(op.OrderID, op.PromotionID, fileOffset)
		}

		// Move to next record
		recordTotalSize := RecordLengthSize + int(recordLength)
		fileOffset += int64(recordTotalSize)
		dataOffset = lengthEnd + int(recordLength)
	}

	// Save the rebuilt index
	if err := hashIndex.Save(indexPath); err != nil {
		return nil, fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	return hashIndex, nil
}
