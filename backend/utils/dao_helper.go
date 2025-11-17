package utils

import (
	"BinaryCRUD/backend/index"
	"strings"
)

// InitializeDAOIndex creates an index path and loads or creates a B+ tree index
// Returns the index path and the loaded/created tree
func InitializeDAOIndex(filePath string) (string, *index.BTree) {
	indexPath := strings.TrimSuffix(filePath, ".bin") + ".idx"

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// If load fails, create new empty tree with default order
		tree = index.NewBTree(DefaultBTreeOrder)
	}

	return indexPath, tree
}
