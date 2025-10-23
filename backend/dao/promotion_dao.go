package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
)

// PromotionDAO handles data access operations for promotions
type PromotionDAO struct {
	filePath string
}

// NewPromotionDAO creates a new PromotionDAO instance
func NewPromotionDAO(filePath string) *PromotionDAO {
	return &PromotionDAO{
		filePath: filePath,
	}
}

// InitializeFile creates and initializes the promotion binary file with header only
// Does not write any records - just creates the file structure
func (dao *PromotionDAO) InitializeFile() error {
	return utils.InitializeBinaryFile(dao.filePath)
}

// Write adds a new promotion to the binary file with auto-increment ID
// Creates and initializes the file if it doesn't exist
// Promotion record format: [ID(4)][UnitSeparator][NameLength(2)][UnitSeparator][Name][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...[RecordSeparator]
// Each item name uses the variable format: [StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
func (dao *PromotionDAO) Write(promotionName string, itemNames []string) error {
	// Initialize file if it doesn't exist
	if err := utils.InitializeBinaryFile(dao.filePath); err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}

	// Get the next ID and increment it atomically
	promotionID, err := utils.GetNextIDAndIncrement(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to get next ID: %w", err)
	}

	// Build the ID field: [ID(4)][UnitSeparator]
	idBytes, err := utils.BuildFixed(4, uint64(promotionID))
	if err != nil {
		return fmt.Errorf("failed to build ID field: %w", err)
	}

	// Build the promotion name field using BuildVariable
	nameBytes, err := utils.BuildVariable(promotionName)
	if err != nil {
		return fmt.Errorf("failed to build promotion name: %w", err)
	}

	// Build the item count field: [ItemCount(2)][UnitSeparator]
	itemCount := uint64(len(itemNames))
	countBytes, err := utils.BuildFixed(2, itemCount)
	if err != nil {
		return fmt.Errorf("failed to build item count: %w", err)
	}

	// Start building the record with ID, name, and count
	recordBytes := append(idBytes, nameBytes...)
	recordBytes = append(recordBytes, countBytes...)

	// Add each item name using BuildVariable
	for i, itemName := range itemNames {
		itemBytes, err := utils.BuildVariable(itemName)
		if err != nil {
			return fmt.Errorf("failed to build item name %d: %w", i, err)
		}
		recordBytes = append(recordBytes, itemBytes...)
	}

	// Append the record to the file
	if err := utils.AppendRecord(dao.filePath, recordBytes); err != nil {
		return fmt.Errorf("failed to append promotion record: %w", err)
	}

	utils.DebugPrint("Successfully wrote promotion [ID:%d] [Name:%s] with %d items", promotionID, promotionName, itemCount)
	return nil
}

// PromotionDTO represents a promotion with its ID, name, and items
type PromotionDTO struct {
	ID    uint32   `json:"id"`
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

// Read reads all promotions from the binary file and returns them with their IDs
func (dao *PromotionDAO) Read() ([]PromotionDTO, error) {
	promotions := []PromotionDTO{}

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return promotions, err
	}

	// Parse each record's bytes to extract the ID, name, item count, and item names
	// Format: [ID(4)][UnitSeparator][NameLength(2)][UnitSeparator][Name][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...
	for _, recordBytes := range records {
		if len(recordBytes) < 8 { // Minimum: 4 bytes ID + 1 sep + 2 bytes name length + 1 sep
			return promotions, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		promotionID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after ID", promotionID)
		}

		// Extract promotion name length (bytes 5-6, little-endian)
		nameLength := uint16(recordBytes[5]) | uint16(recordBytes[6])<<8

		// Verify unit separator after name length at position 7
		if recordBytes[7] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after name length", promotionID)
		}

		// Extract promotion name
		nameStart := 8
		nameEnd := nameStart + int(nameLength)
		if nameEnd > len(recordBytes) {
			return promotions, fmt.Errorf("invalid record ID %d: name length overflow", promotionID)
		}

		promotionName := string(recordBytes[nameStart:nameEnd])
		pos := nameEnd

		// Skip unit separator after name
		if pos < len(recordBytes) && recordBytes[pos] == utils.UnitSeparator {
			pos++
		}

		// Extract item count (2 bytes, little-endian)
		if pos+2 > len(recordBytes) {
			return promotions, fmt.Errorf("invalid record ID %d: missing item count", promotionID)
		}
		itemCount := uint16(recordBytes[pos]) | uint16(recordBytes[pos+1])<<8
		pos += 2

		// Verify unit separator after count
		if pos >= len(recordBytes) || recordBytes[pos] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after item count", promotionID)
		}
		pos++

		// Parse items
		items := []string{}
		for i := 0; i < int(itemCount); i++ {
			if pos+3 > len(recordBytes) { // Need at least length (2 bytes) + separator
				return promotions, fmt.Errorf("invalid record ID %d: incomplete item %d", promotionID, i)
			}

			// Extract item length (2 bytes, little-endian)
			itemLength := uint16(recordBytes[pos]) | uint16(recordBytes[pos+1])<<8
			pos += 2

			// Verify unit separator
			if recordBytes[pos] != utils.UnitSeparator {
				return promotions, fmt.Errorf("invalid record ID %d: missing separator before item %d", promotionID, i)
			}
			pos++

			// Extract item content
			contentEnd := pos + int(itemLength)
			if contentEnd > len(recordBytes) {
				return promotions, fmt.Errorf("invalid record ID %d: item %d content overflow", promotionID, i)
			}

			itemName := string(recordBytes[pos:contentEnd])
			items = append(items, itemName)
			pos = contentEnd

			// Skip unit separator after content
			if pos < len(recordBytes) && recordBytes[pos] == utils.UnitSeparator {
				pos++
			}
		}

		promotions = append(promotions, PromotionDTO{
			ID:    promotionID,
			Name:  promotionName,
			Items: items,
		})
	}

	return promotions, nil
}

// ReadByID reads a single promotion by its ID
func (dao *PromotionDAO) ReadByID(promotionID uint32) (*PromotionDTO, error) {
	promotions, err := dao.Read()
	if err != nil {
		return nil, err
	}

	for _, promotion := range promotions {
		if promotion.ID == promotionID {
			return &promotion, nil
		}
	}

	return nil, fmt.Errorf("promotion with ID %d not found", promotionID)
}

// Print returns a formatted string representation of the binary file contents
func (dao *PromotionDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
