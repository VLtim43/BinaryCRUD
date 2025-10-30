package dao

import (
	"BinaryCRUD/backend/utils"
	"encoding/binary"
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
func (dao *OrderDAO) InitializeFile() error {
	return utils.InitializeBinaryFile(dao.filePath)
}

// Write adds a new order to the binary file with auto-increment ID
// Order record format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...[RecordSeparator]
// Each item name uses the variable format: [StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
func (dao *OrderDAO) Write(itemNames []string) error {
	// Initialize file if it doesn't exist
	if err := utils.InitializeBinaryFile(dao.filePath); err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}

	// Open file in read-write mode for the entire operation
	file, err := utils.OpenFileForWrite(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read current header
	header, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Get the ID to use for this record
	orderID := header.NextID

	// Build the ID field: [ID(4)][UnitSeparator]
	idBytes, err := utils.BuildFixed(4, uint64(orderID))
	if err != nil {
		return fmt.Errorf("failed to build ID field: %w", err)
	}

	// Build the tombstone field: [Tombstone(1)][UnitSeparator] - 0x00 = active
	tombstoneBytes, err := utils.BuildFixed(1, 0x00)
	if err != nil {
		return fmt.Errorf("failed to build tombstone field: %w", err)
	}

	// Build the item count field: [ItemCount(2)][UnitSeparator]
	itemCount := uint64(len(itemNames))
	countBytes, err := utils.BuildFixed(2, itemCount)
	if err != nil {
		return fmt.Errorf("failed to build item count: %w", err)
	}

	// Start building the record with ID, tombstone, and count
	recordBytes := append(idBytes, tombstoneBytes...)
	recordBytes = append(recordBytes, countBytes...)

	// Add each item name using BuildVariable
	for i, itemName := range itemNames {
		nameBytes, err := utils.BuildVariable(itemName)
		if err != nil {
			return fmt.Errorf("failed to build item name %d: %w", i, err)
		}
		recordBytes = append(recordBytes, nameBytes...)
	}

	// Seek to end of file to append the new record
	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	// Write the record bytes
	if _, err := file.Write(recordBytes); err != nil {
		return fmt.Errorf("failed to write record data: %w", err)
	}

	// Write record separator to mark end of record
	if _, err := file.Write([]byte{utils.RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	// Update header with new counts
	header.EntryCount++
	header.NextID++

	// Write updated header using the same file handle
	if err := utils.WriteHeader(file, header); err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	utils.DebugPrint("Successfully wrote order [ID:%d] with %d items", orderID, itemCount)
	return nil
}

type OrderDTO struct {
	ID    uint32   `json:"id"`
	Items []string `json:"items"`
}

// Read reads all orders from the binary file and returns them with their IDs
// Skips records marked as deleted (tombstone = 0x01)
func (dao *OrderDAO) Read() ([]OrderDTO, error) {
	orders := []OrderDTO{}

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return orders, err
	}

	// Parse each record's bytes to extract the ID, item count, and item names
	// Format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...
	for _, recordBytes := range records {
		if len(recordBytes) < 11 { // Minimum: 4 bytes ID + 1 sep + 1 tombstone + 1 sep + 2 bytes count + 1 sep + 0 items + 1 sep
			return orders, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		orderID := binary.LittleEndian.Uint32(recordBytes[0:4])

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return orders, fmt.Errorf("invalid record ID %d: missing unit separator after ID", orderID)
		}

		// Extract tombstone flag (byte 5)
		tombstone := recordBytes[5]

		// Verify unit separator after tombstone at position 6
		if recordBytes[6] != utils.UnitSeparator {
			return orders, fmt.Errorf("invalid record ID %d: missing unit separator after tombstone", orderID)
		}

		// Skip deleted records
		if tombstone == 0x01 {
			continue
		}

		itemCount := binary.LittleEndian.Uint16(recordBytes[7:9])

		// Verify unit separator after count at position 9
		if recordBytes[9] != utils.UnitSeparator {
			return orders, fmt.Errorf("invalid record ID %d: missing unit separator after count", orderID)
		}

		// Parse items starting from position 10
		items := []string{}
		pos := 10

		for i := 0; i < int(itemCount); i++ {
			if pos+3 > len(recordBytes) { // Need at least length (2 bytes) + separator
				return orders, fmt.Errorf("invalid record ID %d: incomplete item %d", orderID, i)
			}

			itemLength := binary.LittleEndian.Uint16(recordBytes[pos : pos+2])
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

// Delete marks an order as deleted by setting its tombstone flag to 0x01
// Returns the number of items in the deleted order
func (dao *OrderDAO) Delete(orderID uint32) (int, error) {
	// Open file for reading and writing
	file, err := utils.OpenBinaryFile(dao.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header to skip it
	_, err = utils.ReadHeader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	currentOffset := int64(utils.HeaderSize)

	// Read records sequentially to find the one with matching ID
	buf := make([]byte, 1)
	for {
		recordStartOffset := currentOffset

		// Read until record separator or EOF
		recordBytes := []byte{}
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					if len(recordBytes) == 0 {
						// Reached end without finding the record
						return 0, fmt.Errorf("order with ID %d not found", orderID)
					}
					return 0, fmt.Errorf("unexpected EOF in record")
				}
				return 0, fmt.Errorf("failed to read byte: %w", err)
			}
			if n == 0 {
				return 0, fmt.Errorf("order with ID %d not found", orderID)
			}

			currentOffset++

			if buf[0] == utils.RecordSeparator {
				break
			}

			recordBytes = append(recordBytes, buf[0])
		}

		// Parse the record to check ID
		if len(recordBytes) < 11 {
			continue
		}

		// Extract ID
		recordID := binary.LittleEndian.Uint32(recordBytes[0:4])

		// Check if this is the record we're looking for
		if recordID == orderID {
			// Verify separator after ID
			if recordBytes[4] != utils.UnitSeparator {
				return 0, fmt.Errorf("invalid record: missing separator after ID")
			}

			// Check tombstone status
			tombstone := recordBytes[5]
			if tombstone == 0x01 {
				return 0, fmt.Errorf("order with ID %d is already deleted", orderID)
			}

			// Verify separator after tombstone
			if recordBytes[6] != utils.UnitSeparator {
				return 0, fmt.Errorf("invalid record: missing separator after tombstone")
			}

			// Extract item count before deletion
			itemCount := binary.LittleEndian.Uint16(recordBytes[7:9])

			// Open file for writing (need a new file handle for O_RDWR)
			writeFile, err := utils.OpenFileForWrite(dao.filePath)
			if err != nil {
				return 0, fmt.Errorf("failed to open file for writing: %w", err)
			}
			defer writeFile.Close()

			// Seek to tombstone position (recordStartOffset + 4 bytes ID + 1 byte separator = recordStartOffset + 5)
			tombstoneOffset := recordStartOffset + 5
			if _, err := writeFile.Seek(tombstoneOffset, 0); err != nil {
				return 0, fmt.Errorf("failed to seek to tombstone position: %w", err)
			}

			// Write 0x01 to mark as deleted
			if _, err := writeFile.Write([]byte{0x01}); err != nil {
				return 0, fmt.Errorf("failed to write tombstone flag: %w", err)
			}

			// Increment tombstone count in header
			if err := utils.IncrementTombstoneCount(dao.filePath); err != nil {
				utils.DebugPrint("Warning: failed to increment tombstone count: %v", err)
			}

			utils.DebugPrint("Successfully deleted order [ID:%d] with %d items", orderID, itemCount)
			return int(itemCount), nil
		}
	}
}

// Print returns a formatted string representation of the binary file contents
func (dao *OrderDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
