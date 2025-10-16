package b_tree

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Index File Format:
// HEADER: [Magic:4bytes "BIDX"] [Version:1byte] [Order:4bytes] [EntryCount:4bytes]
// ENTRIES: [Key:4bytes] [Value:8bytes] ... (sorted by key)
//
// We serialize only the key-value pairs from leaf nodes, not the tree structure.
// On load, we rebuild the tree by inserting entries in order.

const (
	IndexMagic   = "BIDX" // Binary InDeX
	IndexVersion = 1
)

// SaveToFile serializes the B+ tree to a binary file
func (t *BPlusTree) SaveToFile(filename string) error {
	// Get all entries (sorted)
	entries := t.GetAllLeafEntries()

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	// Write magic number
	if _, err := file.WriteString(IndexMagic); err != nil {
		return fmt.Errorf("failed to write magic: %w", err)
	}

	// Write version
	if err := binary.Write(file, binary.LittleEndian, uint8(IndexVersion)); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	// Write order
	if err := binary.Write(file, binary.LittleEndian, uint32(t.Order)); err != nil {
		return fmt.Errorf("failed to write order: %w", err)
	}

	// Write entry count
	entryCount := uint32(len(entries))
	if err := binary.Write(file, binary.LittleEndian, entryCount); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		key := entry[0].(uint32)
		value := entry[1].(int64)

		// Write key (4 bytes)
		if err := binary.Write(file, binary.LittleEndian, key); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}

		// Write value (8 bytes)
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write value: %w", err)
		}
	}

	return nil
}

// LoadFromFile deserializes a B+ tree from a binary file
func LoadFromFile(filename string) (*BPlusTree, error) {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	// Read and verify magic number
	magic := make([]byte, 4)
	if _, err := io.ReadFull(file, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if string(magic) != IndexMagic {
		return nil, fmt.Errorf("invalid index file: bad magic number")
	}

	// Read version
	var version uint8
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version != IndexVersion {
		return nil, fmt.Errorf("unsupported index version: %d", version)
	}

	// Read order
	var order uint32
	if err := binary.Read(file, binary.LittleEndian, &order); err != nil {
		return nil, fmt.Errorf("failed to read order: %w", err)
	}

	// Read entry count
	var entryCount uint32
	if err := binary.Read(file, binary.LittleEndian, &entryCount); err != nil {
		return nil, fmt.Errorf("failed to read entry count: %w", err)
	}

	// Create new tree
	tree := NewBPlusTree(int(order))

	// Read and insert entries
	for i := uint32(0); i < entryCount; i++ {
		// Read key
		var key uint32
		if err := binary.Read(file, binary.LittleEndian, &key); err != nil {
			return nil, fmt.Errorf("failed to read key at entry %d: %w", i, err)
		}

		// Read value
		var value int64
		if err := binary.Read(file, binary.LittleEndian, &value); err != nil {
			return nil, fmt.Errorf("failed to read value at entry %d: %w", i, err)
		}

		// Insert into tree (rebuilds structure)
		tree.Insert(key, value)
	}

	return tree, nil
}

// IndexFileExists checks if an index file exists
func IndexFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
