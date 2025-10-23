package dao

import (
	"BinaryCRUD/backend/utils"
)

// OrderDAO handles data access operations for orders
type OrderDAO struct {
	filePath string
}

// NewOrderDAO creates a new OrderDAO instance
func NewOrderDAO(filePath string) *OrderDAO {
	return &OrderDAO{
		filePath: filePath,
	}
}

// InitializeFile creates and initializes the order binary file with header only
// Does not write any records - just creates the file structure
func (dao *OrderDAO) InitializeFile() error {
	return utils.InitializeBinaryFile(dao.filePath)
}
