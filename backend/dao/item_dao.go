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

// Write adds a new item to the binary file with auto-increment ID
// Creates and initializes the file if it doesn't exist
// Item record format: [ID(4)][UnitSeparator][StringLength(2)][UnitSeparator][StringContent][UnitSeparator][RecordSeparator]
func (dao *ItemDAO) Write(itemName string) error {
	// Initialize file if it doesn't exist
	if err := utils.InitializeBinaryFile(dao.filePath); err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}

	// Get the next ID and increment it atomically
	itemID, err := utils.GetNextIDAndIncrement(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to get next ID: %w", err)
	}

	// Build the ID field: [ID(4)][UnitSeparator]
	idBytes, err := utils.BuildFixed(4, uint64(itemID))
	if err != nil {
		return fmt.Errorf("failed to build ID field: %w", err)
	}

	// Build the item name field: [StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
	nameBytes, err := utils.BuildVariable(itemName)
	if err != nil {
		return fmt.Errorf("failed to build item name: %w", err)
	}

	// Combine ID and name bytes
	recordBytes := append(idBytes, nameBytes...)

	// Append the record to the file
	if err := utils.AppendRecord(dao.filePath, recordBytes); err != nil {
		return fmt.Errorf("failed to append item record: %w", err)
	}

	utils.DebugPrint("Successfully wrote item [ID:%d]: \"%s\"", itemID, itemName)
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

	// Parse each record's bytes to extract the ID and item name
	// Format: [ID(4)][UnitSeparator][StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
	for _, recordBytes := range records {
		if len(recordBytes) < 8 { // Minimum: 4 bytes ID + 1 sep + 2 bytes length + 1 sep
			return items, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		itemID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after ID", itemID)
		}

		// Extract length (bytes 5-6, little-endian)
		length := uint16(recordBytes[5]) | uint16(recordBytes[6])<<8

		// Verify unit separator after length at position 7
		if recordBytes[7] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after length", itemID)
		}

		// Extract content (after second separator)
		contentStart := 8
		contentEnd := contentStart + int(length)

		if contentEnd > len(recordBytes) {
			return items, fmt.Errorf("invalid record ID %d: content length mismatch", itemID)
		}

		// Verify final unit separator
		if contentEnd < len(recordBytes) && recordBytes[contentEnd] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing final unit separator", itemID)
		}

		itemName := string(recordBytes[contentStart:contentEnd])
		items[itemID] = itemName
	}

	return items, nil
}

// Print returns a formatted string representation of the binary file contents
func (dao *ItemDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
