package utils

import (
	"BinaryCRUD/backend/index"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// IndexPathFromBinFile extracts index path from a .bin file path
// e.g., "data/items.bin" -> "data/indexes/items.idx"
func IndexPathFromBinFile(filePath string) string {
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ".bin")
	return filepath.Join("data", "indexes", baseName+".idx")
}

// RebuildFunc is a function type for index rebuilding
type RebuildFunc func(binFilePath, indexPath string) error

// initializeBTreeIndex is a generic helper for B+ tree index initialization
func initializeBTreeIndex(filePath string, rebuildFn func(string, string) (*index.BTree, error)) (string, *index.BTree) {
	indexPath := IndexPathFromBinFile(filePath)

	tree, err := index.Load(indexPath)
	if err != nil {
		log.Printf("Index load failed for %s, rebuilding from data file...", indexPath)
		tree, err = rebuildFn(filePath, indexPath)
		if err != nil {
			log.Printf("Index rebuild failed: %v, creating empty tree", err)
			tree = index.NewBTree(DefaultBTreeOrder)
		} else {
			log.Printf("Index rebuilt successfully for %s", indexPath)
		}
	} else {
		os.Remove(indexPath + ".tmp")
	}

	return indexPath, tree
}

// InitializeDAOIndex creates an index path and loads or creates a B+ tree index for items
// Returns the index path and the loaded/created tree
// Index files are stored in data/indexes/ directory
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeDAOIndex(filePath string) (string, *index.BTree) {
	return initializeBTreeIndex(filePath, RebuildBTreeIndex)
}

// InitializeCollectionDAOIndex creates an index for collections (orders/promotions)
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeCollectionDAOIndex(filePath string) (string, *index.BTree) {
	return initializeBTreeIndex(filePath, RebuildCollectionBTreeIndex)
}

// InitializeOrderPromotionIndex creates an extensible hash index for order-promotion relationships
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeOrderPromotionIndex(filePath string, bucketSize int) (string, *index.ExtensibleHash) {
	indexPath := IndexPathFromBinFile(filePath)

	hashIndex, err := index.LoadExtensibleHash(indexPath)
	if err != nil {
		log.Printf("Hash index load failed for %s, rebuilding from data file...", indexPath)
		hashIndex, err = RebuildExtensibleHashIndex(filePath, indexPath, bucketSize)
		if err != nil {
			log.Printf("Hash index rebuild failed: %v, creating empty hash", err)
			hashIndex = index.NewExtensibleHash(bucketSize)
		} else {
			log.Printf("Hash index rebuilt successfully for %s", indexPath)
		}
	} else {
		os.Remove(indexPath + ".tmp")
	}

	return indexPath, hashIndex
}

// DeleteFromBTreeIndex handles the common delete pattern for B+ tree indexed DAOs.
// It removes from index, saves the index, then soft deletes the entry.
// Returns a formatted error with the entity name if something fails.
func DeleteFromBTreeIndex(tree *index.BTree, indexPath, filePath string, id uint64, entityName string) error {
	// Remove from index first
	err := tree.Delete(id)
	if err != nil {
		return fmt.Errorf("%s not found: %w", entityName, err)
	}

	// Save updated index
	err = tree.Save(indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Soft delete the entry in the file
	return SoftDeleteByID(filePath, id, nil, nil)
}

// CombineBytes efficiently combines multiple byte slices into one.
// Pre-allocates the exact capacity needed to avoid reallocations.
func CombineBytes(slices ...[]byte) []byte {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}
	result := make([]byte, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
