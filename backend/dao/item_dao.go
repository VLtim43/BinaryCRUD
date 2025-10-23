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

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return items, err
	}

	// Parse each record's bytes to extract the item name
	for id, recordBytes := range records {
		// Parse the variable-length string from the record bytes
		// Format: [Length(2)][UnitSeparator][Content][UnitSeparator]
		if len(recordBytes) < 4 { // Minimum: 2 bytes length + 2 separators
			return items, fmt.Errorf("invalid record at index %d: too short", id)
		}

		// Extract length (first 2 bytes, little-endian)
		length := uint16(recordBytes[0]) | uint16(recordBytes[1])<<8

		// Verify first unit separator at position 2
		if recordBytes[2] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record at index %d: missing first unit separator", id)
		}

		// Extract content (after length and separator, before final separator)
		contentStart := 3
		contentEnd := contentStart + int(length)

		if contentEnd > len(recordBytes) {
			return items, fmt.Errorf("invalid record at index %d: content length mismatch", id)
		}

		itemName := string(recordBytes[contentStart:contentEnd])
		items[id] = itemName
	}

	return items, nil
}

// Print returns a formatted string representation of the binary file contents
func (dao *ItemDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
