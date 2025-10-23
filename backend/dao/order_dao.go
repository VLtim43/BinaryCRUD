package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
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

// Write adds a new order to the binary file with auto-increment ID
// Creates and initializes the file if it doesn't exist
// Order record format: [ID(4)][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...[RecordSeparator]
// Each item name uses the variable format: [StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
func (dao *OrderDAO) Write(itemNames []string) error {
	// Initialize file if it doesn't exist
	if err := utils.InitializeBinaryFile(dao.filePath); err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}

	// Get the next ID and increment it atomically
	orderID, err := utils.GetNextIDAndIncrement(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to get next ID: %w", err)
	}

	// Build the ID field: [ID(4)][UnitSeparator]
	idBytes, err := utils.BuildFixed(4, uint64(orderID))
	if err != nil {
		return fmt.Errorf("failed to build ID field: %w", err)
	}

	// Build the item count field: [ItemCount(2)][UnitSeparator]
	itemCount := uint64(len(itemNames))
	countBytes, err := utils.BuildFixed(2, itemCount)
	if err != nil {
		return fmt.Errorf("failed to build item count: %w", err)
	}

	// Start building the record with ID and count
	recordBytes := append(idBytes, countBytes...)

	// Add each item name using BuildVariable
	for i, itemName := range itemNames {
		nameBytes, err := utils.BuildVariable(itemName)
		if err != nil {
			return fmt.Errorf("failed to build item name %d: %w", i, err)
		}
		recordBytes = append(recordBytes, nameBytes...)
	}

	// Append the record to the file
	if err := utils.AppendRecord(dao.filePath, recordBytes); err != nil {
		return fmt.Errorf("failed to append order record: %w", err)
	}

	utils.DebugPrint("Successfully wrote order [ID:%d] with %d items", orderID, itemCount)
	return nil
}
