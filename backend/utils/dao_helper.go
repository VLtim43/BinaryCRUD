package utils

import (
	"BinaryCRUD/backend/index"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// InitializeDAOIndex creates an index path and loads or creates a B+ tree index for items
// Returns the index path and the loaded/created tree
// Index files are stored in data/indexes/ directory
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeDAOIndex(filePath string) (string, *index.BTree) {
	// Extract just the filename without extension
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ".bin")

	// Put index in data/indexes/ directory
	indexPath := filepath.Join("data", "indexes", baseName+".idx")

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// Index missing or corrupted - rebuild from .bin file
		log.Printf("Index load failed for %s, rebuilding from data file...", indexPath)
		tree, err = RebuildBTreeIndex(filePath, indexPath)
		if err != nil {
			log.Printf("Index rebuild failed: %v, creating empty tree", err)
			tree = index.NewBTree(DefaultBTreeOrder)
		} else {
			log.Printf("Index rebuilt successfully for %s", indexPath)
		}
	} else {
		// Clean up any leftover temp files from interrupted saves
		os.Remove(indexPath + ".tmp")
	}

	return indexPath, tree
}

// InitializeCollectionDAOIndex creates an index for collections (orders/promotions)
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeCollectionDAOIndex(filePath string) (string, *index.BTree) {
	// Extract just the filename without extension
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ".bin")

	// Put index in data/indexes/ directory
	indexPath := filepath.Join("data", "indexes", baseName+".idx")

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// Index missing or corrupted - rebuild from .bin file
		log.Printf("Index load failed for %s, rebuilding from data file...", indexPath)
		tree, err = RebuildCollectionBTreeIndex(filePath, indexPath)
		if err != nil {
			log.Printf("Index rebuild failed: %v, creating empty tree", err)
			tree = index.NewBTree(DefaultBTreeOrder)
		} else {
			log.Printf("Index rebuilt successfully for %s", indexPath)
		}
	} else {
		// Clean up any leftover temp files from interrupted saves
		os.Remove(indexPath + ".tmp")
	}

	return indexPath, tree
}

// InitializeOrderPromotionIndex creates an extensible hash index for order-promotion relationships
// If index is missing or corrupted, it will be rebuilt from the .bin file
func InitializeOrderPromotionIndex(filePath string, bucketSize int) (string, *index.ExtensibleHash) {
	// Extract just the filename without extension
	baseName := filepath.Base(filePath)
	baseName = strings.TrimSuffix(baseName, ".bin")

	// Put index in data/indexes/ directory
	indexPath := filepath.Join("data", "indexes", baseName+".idx")

	// Try to load existing index
	hashIndex, err := index.LoadExtensibleHash(indexPath)
	if err != nil {
		// Index missing or corrupted - rebuild from .bin file
		log.Printf("Hash index load failed for %s, rebuilding from data file...", indexPath)
		hashIndex, err = RebuildExtensibleHashIndex(filePath, indexPath, bucketSize)
		if err != nil {
			log.Printf("Hash index rebuild failed: %v, creating empty hash", err)
			hashIndex = index.NewExtensibleHash(bucketSize)
		} else {
			log.Printf("Hash index rebuilt successfully for %s", indexPath)
		}
	} else {
		// Clean up any leftover temp files from interrupted saves
		os.Remove(indexPath + ".tmp")
	}

	return indexPath, hashIndex
}
