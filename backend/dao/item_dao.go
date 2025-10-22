package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
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
	utils.DebugPrint("Writing item", itemName)

	// Build the record bytes using BuildVariable
	recordBytes, err := utils.BuildVariable(itemName)
	if err != nil {
		return fmt.Errorf("failed to build item record: %w", err)
	}

	// Append the record to the file
	if err := utils.AppendRecord(dao.filePath, recordBytes); err != nil {
		return fmt.Errorf("failed to append item record: %w", err)
	}

	utils.DebugPrint("Successfully wrote item", itemName)
	return nil
}

// Print returns a formatted string representation of the binary file contents
func (dao *ItemDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
