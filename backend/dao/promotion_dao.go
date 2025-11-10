package dao

import (
	"BinaryCRUD/backend/index"
)

// PromotionDAO wraps CollectionDAO for promotions
type PromotionDAO struct {
	*CollectionDAO
}

// NewPromotionDAO creates a DAO for promotions.bin with B+ Tree index
func NewPromotionDAO(filePath string) *PromotionDAO {
	indexPath := filePath[:len(filePath)-4] + ".idx" // Replace .bin with .idx

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// If load fails, create new empty tree
		tree = index.NewBTree(4)
	}

	return &PromotionDAO{
		CollectionDAO: &CollectionDAO{
			filePath:  filePath,
			indexPath: indexPath,
			tree:      tree,
		},
	}
}

// GetIndexTree returns the B+ tree index
func (dao *PromotionDAO) GetIndexTree() *index.BTree {
	return dao.tree
}
