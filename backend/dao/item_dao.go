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
	// Build the record bytes using BuildVariable
	recordBytes, err := utils.BuildVariable(itemName)
	if err != nil {
		return fmt.Errorf("failed to build item record: %w", err)
	}

	// Append the record to the file
	if err := utils.AppendRecord(dao.filePath, recordBytes); err != nil {
		return fmt.Errorf("failed to append item record: %w", err)
	}

	utils.DebugPrint("Successfully wrote item: \"%s\"", itemName)
	return nil
}

// Read reads all items from the binary file and returns them with their IDs
func (dao *ItemDAO) Read() (map[uint32]string, error) {
	items := make(map[uint32]string)

	// Check if file exists
	file, err := utils.OpenBinaryFile(dao.filePath)
	if err != nil {
		return items, err
	}
	defer file.Close()

	// Read header to get entry count
	header, err := utils.ReadHeader(file)
	if err != nil {
		return items, fmt.Errorf("failed to read header: %w", err)
	}

	// Read each item
	for i := uint32(0); i < header.EntryCount; i++ {
		itemName, err := utils.ReadVariable(file)
		if err != nil {
			return items, fmt.Errorf("failed to read item at index %d: %w", i, err)
		}

		// Read record separator
		sep := make([]byte, 1)
		if _, err := file.Read(sep); err != nil {
			return items, fmt.Errorf("failed to read record separator at index %d: %w", i, err)
		}

		items[i] = itemName
	}

	return items, nil
}

// Print returns a formatted string representation of the binary file contents
func (dao *ItemDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
