package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
)

// PromotionDAO wraps CollectionDAO for promotions
type PromotionDAO struct {
	*CollectionDAO
}

// NewPromotionDAO creates a DAO for promotions.bin with B+ Tree index
func NewPromotionDAO(filePath string) *PromotionDAO {
	indexPath, tree := utils.InitializeDAOIndex(filePath)

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
