package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
)

// ItemDAO handles data access operations for items
type ItemDAO struct {
	filePath string
}

// NewItemDAO creates a new ItemDAO instance
func NewItemDAO(filePath string) *ItemDAO {
	return &ItemDAO{
		filePath: filePath,
	}
}

// Write adds a new item to the binary file
// Creates and initializes the file if it doesn't exist
// Item record format: [ItemName(variable)][RecordSeparator]
func (dao *ItemDAO) Write(itemName string) error {
	fmt.Printf("[ItemDAO] Writing item: %s\n", itemName)

	// Use the generic AppendRecord function
	err := utils.AppendRecord(dao.filePath, func(file *os.File) error {
		// Write item name (variable-length with size prefix and unit separator)
		return utils.WriteVariable(file, itemName)
	})

	if err != nil {
		return fmt.Errorf("failed to append item record: %w", err)
	}

	fmt.Printf("[ItemDAO] Successfully wrote item: %s\n", itemName)
	return nil
}
