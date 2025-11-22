package index

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// Save writes the tree to a file
func (t *BTree) Save(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	// Get all entries
	entries := t.GetAll()

	// Write count
	count := uint64(len(entries))
	if err := binary.Write(file, binary.BigEndian, count); err != nil {
		return fmt.Errorf("failed to write count: %w", err)
	}

	// Write each entry
	for id, offset := range entries {
		if err := binary.Write(file, binary.BigEndian, id); err != nil {
			return fmt.Errorf("failed to write id: %w", err)
		}
		if err := binary.Write(file, binary.BigEndian, offset); err != nil {
			return fmt.Errorf("failed to write offset: %w", err)
		}
	}

	// Sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	return nil
}

// Load reads the tree from a file
func Load(path string) (*BTree, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty tree
			return NewBTree(4), nil
		}
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	tree := NewBTree(4)

	// Read count
	var count uint64
	if err := binary.Read(file, binary.BigEndian, &count); err != nil {
		return nil, fmt.Errorf("failed to read count: %w", err)
	}

	// Read each entry
	for i := uint64(0); i < count; i++ {
		var id uint64
		var offset int64

		if err := binary.Read(file, binary.BigEndian, &id); err != nil {
			return nil, fmt.Errorf("failed to read id: %w", err)
		}
		if err := binary.Read(file, binary.BigEndian, &offset); err != nil {
			return nil, fmt.Errorf("failed to read offset: %w", err)
		}

		if err := tree.Insert(id, offset); err != nil {
			return nil, fmt.Errorf("failed to insert: %w", err)
		}
	}

	return tree, nil
}
