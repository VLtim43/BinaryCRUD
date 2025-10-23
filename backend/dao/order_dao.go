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

// OrderDTO represents an order with its ID and items
type OrderDTO struct {
	ID    uint32   `json:"id"`
	Items []string `json:"items"`
}

// Read reads all orders from the binary file and returns them with their IDs
func (dao *OrderDAO) Read() ([]OrderDTO, error) {
	orders := []OrderDTO{}

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return orders, err
	}

	// Parse each record's bytes to extract the ID, item count, and item names
	// Format: [ID(4)][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...
	for _, recordBytes := range records {
		if len(recordBytes) < 8 { // Minimum: 4 bytes ID + 1 sep + 2 bytes count + 1 sep
			return orders, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		orderID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return orders, fmt.Errorf("invalid record ID %d: missing unit separator after ID", orderID)
		}

		// Extract item count (bytes 5-6, little-endian)
		itemCount := uint16(recordBytes[5]) | uint16(recordBytes[6])<<8

		// Verify unit separator after count at position 7
		if recordBytes[7] != utils.UnitSeparator {
			return orders, fmt.Errorf("invalid record ID %d: missing unit separator after count", orderID)
		}

		// Parse items starting from position 8
		items := []string{}
		pos := 8

		for i := 0; i < int(itemCount); i++ {
			if pos+3 > len(recordBytes) { // Need at least length (2 bytes) + separator
				return orders, fmt.Errorf("invalid record ID %d: incomplete item %d", orderID, i)
			}

			// Extract item length (2 bytes, little-endian)
			itemLength := uint16(recordBytes[pos]) | uint16(recordBytes[pos+1])<<8
			pos += 2

			// Verify unit separator
			if recordBytes[pos] != utils.UnitSeparator {
				return orders, fmt.Errorf("invalid record ID %d: missing separator before item %d", orderID, i)
			}
			pos++

			// Extract item content
			contentEnd := pos + int(itemLength)
			if contentEnd > len(recordBytes) {
				return orders, fmt.Errorf("invalid record ID %d: item %d content overflow", orderID, i)
			}

			itemName := string(recordBytes[pos:contentEnd])
			items = append(items, itemName)
			pos = contentEnd

			// Skip unit separator after content
			if pos < len(recordBytes) && recordBytes[pos] == utils.UnitSeparator {
				pos++
			}
		}

		orders = append(orders, OrderDTO{
			ID:    orderID,
			Items: items,
		})
	}

	return orders, nil
}

// ReadByID reads a single order by its ID
func (dao *OrderDAO) ReadByID(orderID uint32) (*OrderDTO, error) {
	orders, err := dao.Read()
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		if order.ID == orderID {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("order with ID %d not found", orderID)
}

// Print returns a formatted string representation of the binary file contents
func (dao *OrderDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
