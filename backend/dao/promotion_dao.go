package dao

import (
	"BinaryCRUD/backend/utils"
	"encoding/binary"
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
func (dao *PromotionDAO) InitializeFile() error {
	return utils.InitializeBinaryFile(dao.filePath)
}

// Write adds a new promotion to the binary file with auto-increment ID
// Creates and initializes the file if it doesn't exist
// Promotion record format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][NameLength(2)][UnitSeparator][Name][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...[RecordSeparator]
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

	// Build the tombstone field: [Tombstone(1)][UnitSeparator] - 0x00 = active
	tombstoneBytes, err := utils.BuildFixed(1, 0x00)
	if err != nil {
		return fmt.Errorf("failed to build tombstone field: %w", err)
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

	// Start building the record with ID, tombstone, name, and count
	recordBytes := append(idBytes, tombstoneBytes...)
	recordBytes = append(recordBytes, nameBytes...)
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

type PromotionDTO struct {
	ID    uint32   `json:"id"`
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

// Read reads all promotions from the binary file and returns them with their IDs
// Skips records marked as deleted (tombstone = 0x01)
func (dao *PromotionDAO) Read() ([]PromotionDTO, error) {
	promotions := []PromotionDTO{}

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return promotions, err
	}

	// Parse each record's bytes to extract the ID, name, item count, and item names
	// Format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][NameLength(2)][UnitSeparator][Name][UnitSeparator][ItemCount(2)][UnitSeparator][Item1Name][Item2Name]...
	for _, recordBytes := range records {
		if len(recordBytes) < 11 { // Minimum: 4 bytes ID + 1 sep + 1 tombstone + 1 sep + 2 bytes name length + 1 sep + 0 name + 1 sep
			return promotions, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		promotionID := binary.LittleEndian.Uint32(recordBytes[0:4])

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after ID", promotionID)
		}

		// Extract tombstone flag (byte 5)
		tombstone := recordBytes[5]

		// Verify unit separator after tombstone at position 6
		if recordBytes[6] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after tombstone", promotionID)
		}

		// Skip deleted records
		if tombstone == 0x01 {
			continue
		}

		// Extract promotion name length (bytes 7-8, little-endian)
		nameLength := binary.LittleEndian.Uint16(recordBytes[7:9])

		// Verify unit separator after name length at position 9
		if recordBytes[9] != utils.UnitSeparator {
			return promotions, fmt.Errorf("invalid record ID %d: missing unit separator after name length", promotionID)
		}

		// Extract promotion name
		nameStart := 10
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
		itemCount := binary.LittleEndian.Uint16(recordBytes[pos : pos+2])
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
			itemLength := binary.LittleEndian.Uint16(recordBytes[pos : pos+2])
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

// Delete marks a promotion as deleted by setting its tombstone flag to 0x01
// Uses sequential search to locate the record
func (dao *PromotionDAO) Delete(promotionID uint32) error {
	// Open file for reading and writing
	file, err := utils.OpenBinaryFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header to skip it
	_, err = utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Track current offset (start after header)
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
						return fmt.Errorf("promotion with ID %d not found", promotionID)
					}
					return fmt.Errorf("unexpected EOF in record")
				}
				return fmt.Errorf("failed to read byte: %w", err)
			}
			if n == 0 {
				return fmt.Errorf("promotion with ID %d not found", promotionID)
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
		if recordID == promotionID {
			// Verify separator after ID
			if recordBytes[4] != utils.UnitSeparator {
				return fmt.Errorf("invalid record: missing separator after ID")
			}

			// Check tombstone status
			tombstone := recordBytes[5]
			if tombstone == 0x01 {
				return fmt.Errorf("promotion with ID %d is already deleted", promotionID)
			}

			// Open file for writing (need a new file handle for O_RDWR)
			writeFile, err := utils.OpenBinaryFile(dao.filePath)
			if err != nil {
				return fmt.Errorf("failed to open file for writing: %w", err)
			}
			defer writeFile.Close()

			// Seek to tombstone position (recordStartOffset + 4 bytes ID + 1 byte separator = recordStartOffset + 5)
			tombstoneOffset := recordStartOffset + 5
			if _, err := writeFile.Seek(tombstoneOffset, 0); err != nil {
				return fmt.Errorf("failed to seek to tombstone position: %w", err)
			}

			// Write 0x01 to mark as deleted
			if _, err := writeFile.Write([]byte{0x01}); err != nil {
				return fmt.Errorf("failed to write tombstone flag: %w", err)
			}

			// Increment tombstone count in header
			if err := utils.IncrementTombstoneCount(dao.filePath); err != nil {
				utils.DebugPrint("Warning: failed to increment tombstone count: %v", err)
			}

			utils.DebugPrint("Successfully deleted promotion [ID:%d]", promotionID)
			return nil
		}
	}
}

// Print returns a formatted string representation of the binary file contents
func (dao *PromotionDAO) Print() (string, error) {
	return utils.PrintBinaryFile(dao.filePath)
}
