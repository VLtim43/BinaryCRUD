package dao

import (
	"BinaryCRUD/backend/index"
)

// OrderDAO wraps CollectionDAO for orders
type OrderDAO struct {
	*CollectionDAO
}

// NewOrderDAO creates a DAO for orders.bin with B+ Tree index
func NewOrderDAO(filePath string) *OrderDAO {
	indexPath := filePath[:len(filePath)-4] + ".idx" // Replace .bin with .idx

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// If load fails, create new empty tree
		tree = index.NewBTree(4)
	}

	return &OrderDAO{
		CollectionDAO: &CollectionDAO{
			filePath:  filePath,
			indexPath: indexPath,
			tree:      tree,
		},
	}
}

// GetIndexTree returns the B+ tree index
func (dao *OrderDAO) GetIndexTree() *index.BTree {
	return dao.tree
}
