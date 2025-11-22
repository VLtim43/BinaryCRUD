package utils

import (
	"BinaryCRUD/backend/index"
	"path/filepath"
	"strings"
)

// InitializeDAOIndex creates an index path and loads or creates a B+ tree index
// Returns the index path and the loaded/created tree
// Index files are stored in data/indexes/ directory
func InitializeDAOIndex(filePath string) (string, *index.BTree) {
	// Extract just the filename without extension
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ".bin")

	// Put index in data/indexes/ directory
	indexPath := filepath.Join("data", "indexes", baseName+".idx")

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// If load fails, create new empty tree with default order
		tree = index.NewBTree(DefaultBTreeOrder)
	}

	return indexPath, tree
}
