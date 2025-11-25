package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
)

// OrderDAO wraps CollectionDAO for orders
type OrderDAO struct {
	*CollectionDAO
}

// NewOrderDAO creates a DAO for orders.bin with B+ Tree index
func NewOrderDAO(filePath string) *OrderDAO {
	indexPath, tree := utils.InitializeCollectionDAOIndex(filePath)

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
