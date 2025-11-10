package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"sync"
)

// OrderPromotion represents the N:N relationship between Orders and Promotions
type OrderPromotion struct {
	OrderID     uint64
	PromotionID uint64
}

type OrderPromotionDAO struct {
	filePath string
	mu       sync.Mutex
}

// NewOrderPromotionDAO creates a DAO for order_promotions.bin
func NewOrderPromotionDAO(filePath string) *OrderPromotionDAO {
	return &OrderPromotionDAO{filePath: filePath}
}

// ensureFileExists creates the file with empty header if it doesn't exist
func (dao *OrderPromotionDAO) ensureFileExists() error {
	// Check if file already exists
	if _, err := os.Stat(dao.filePath); err == nil {
		return nil
	}

	// Create the file
	file, err := utils.CreateFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to create order_promotion file: %w", err)
	}
	defer file.Close()

	// Write empty header (0 entities, 0 tombstones, nextId=0)
	header, err := utils.WriteHeader(0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// Write creates a new order-promotion relationship
// Binary format: [ID(2)][tombstone(1)][0x1F][orderID(2)][0x1F][promotionID(2)][0x1E]
func (dao *OrderPromotionDAO) Write(orderID, promotionID uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open order_promotion file: %w", err)
	}
	defer file.Close()

	// Build entry without ID: [0x1F][orderID(2)][0x1F][promotionID(2)]

	// Separator before orderID
	sep1, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Order ID (2 bytes)
	orderIDBytes, err := utils.WriteFixedNumber(utils.IDSize, orderID)
	if err != nil {
		return fmt.Errorf("failed to write order ID: %w", err)
	}

	// Separator before promotionID
	sep2, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Promotion ID (2 bytes)
	promotionIDBytes, err := utils.WriteFixedNumber(utils.IDSize, promotionID)
	if err != nil {
		return fmt.Errorf("failed to write promotion ID: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, sep1...)
	entry = append(entry, orderIDBytes...)
	entry = append(entry, sep2...)
	entry = append(entry, promotionIDBytes...)

	// Append the entry (ID and tombstone auto-assigned, record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return fmt.Errorf("failed to append order_promotion: %w", err)
	}

	return nil
}

// GetByOrderID retrieves all promotions applied to an order
func (dao *OrderPromotionDAO) GetByOrderID(orderID uint64) ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Read all file data
	fileData, err := os.ReadFile(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate header size
	headerSize := (utils.HeaderFieldSize * 3) + 3 // 15 bytes

	if len(fileData) <= headerSize {
		return []*OrderPromotion{}, nil // No entries yet
	}

	// Split by record separator
	recordSeparatorByte := []byte(utils.RecordSeparator)[0]
	entries := make([][]byte, 0)

	entryStart := headerSize
	for i := headerSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, fileData[entryStart:i])
			entryStart = i + 1
		}
	}

	// Parse each entry and filter by orderID
	result := make([]*OrderPromotion, 0)

	for _, entryData := range entries {
		if len(entryData) < utils.IDSize+utils.TombstoneSize {
			continue
		}

		// Read ID
		_, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, 0)
		if err != nil {
			continue
		}

		// Read tombstone
		tombstone := entryData[offset]
		offset += utils.TombstoneSize

		// Skip deleted entries
		if tombstone != 0x00 {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read orderID
		entryOrderID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, _, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Filter by orderID
		if entryOrderID == orderID {
			result = append(result, &OrderPromotion{
				OrderID:     entryOrderID,
				PromotionID: entryPromotionID,
			})
		}
	}

	return result, nil
}

// GetByPromotionID retrieves all orders that have a specific promotion applied
func (dao *OrderPromotionDAO) GetByPromotionID(promotionID uint64) ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Read all file data
	fileData, err := os.ReadFile(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate header size
	headerSize := (utils.HeaderFieldSize * 3) + 3 // 15 bytes

	if len(fileData) <= headerSize {
		return []*OrderPromotion{}, nil
	}

	// Split by record separator
	recordSeparatorByte := []byte(utils.RecordSeparator)[0]
	entries := make([][]byte, 0)

	entryStart := headerSize
	for i := headerSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, fileData[entryStart:i])
			entryStart = i + 1
		}
	}

	// Parse each entry and filter by promotionID
	result := make([]*OrderPromotion, 0)

	for _, entryData := range entries {
		if len(entryData) < utils.IDSize+utils.TombstoneSize {
			continue
		}

		// Read ID
		_, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, 0)
		if err != nil {
			continue
		}

		// Read tombstone
		tombstone := entryData[offset]
		offset += utils.TombstoneSize

		// Skip deleted entries
		if tombstone != 0x00 {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read orderID
		entryOrderID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, _, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Filter by promotionID
		if entryPromotionID == promotionID {
			result = append(result, &OrderPromotion{
				OrderID:     entryOrderID,
				PromotionID: entryPromotionID,
			})
		}
	}

	return result, nil
}

// Delete removes an order-promotion relationship by marking it as deleted
func (dao *OrderPromotionDAO) Delete(orderID, promotionID uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open order_promotion file: %w", err)
	}
	defer file.Close()

	// Read header
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Read all file data
	fileData, err := os.ReadFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate header size
	headerSize := (utils.HeaderFieldSize * 3) + 3

	// Split by record separator
	recordSeparatorByte := []byte(utils.RecordSeparator)[0]
	entries := make([][]byte, 0)
	entryPositions := make([]int64, 0)

	entryStart := headerSize
	for i := headerSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, fileData[entryStart:i])
			entryPositions = append(entryPositions, int64(entryStart))
			entryStart = i + 1
		}
	}

	// Find the entry with matching orderID and promotionID
	for idx, entryData := range entries {
		if len(entryData) < utils.IDSize+utils.TombstoneSize {
			continue
		}

		offset := 0

		// Read ID
		_, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Check tombstone
		tombstone := entryData[offset]
		offset += utils.TombstoneSize

		if tombstone != 0x00 {
			continue // Already deleted
		}

		// Skip unit separator
		offset += 1

		// Read orderID
		entryOrderID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, _, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Check if this is the entry we want to delete
		if entryOrderID == orderID && entryPromotionID == promotionID {
			// Calculate tombstone position
			tombstonePos := entryPositions[idx] + int64(utils.IDSize)

			// Seek to tombstone position
			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			// Write 0x01 to mark as deleted
			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			// Sync to disk
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			// Update header
			err = utils.UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			// Sync header to disk
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync header to disk: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("order-promotion relationship not found (orderID=%d, promotionID=%d)", orderID, promotionID)
}
