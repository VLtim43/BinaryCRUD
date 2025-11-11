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
	return utils.EnsureFileExists(dao.filePath)
}

// Write creates a new order-promotion relationship
// Binary format with composite primary key: [orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)][0x1E]
// The composite key is (orderID, promotionID) - no auto-generated ID
func (dao *OrderPromotionDAO) Write(orderID, promotionID uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Check for duplicate composite key
	exists, err := dao.existsUnlocked(orderID, promotionID)
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if exists {
		return fmt.Errorf("order-promotion relationship already exists (orderID=%d, promotionID=%d)", orderID, promotionID)
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open order_promotion file: %w", err)
	}
	defer file.Close()

	// Read header to update entity count
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Build entry: [orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)][0x1E]

	// Order ID (2 bytes)
	orderIDBytes, err := utils.WriteFixedNumber(utils.IDSize, orderID)
	if err != nil {
		return fmt.Errorf("failed to write order ID: %w", err)
	}

	// Separator
	sep1, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Promotion ID (2 bytes)
	promotionIDBytes, err := utils.WriteFixedNumber(utils.IDSize, promotionID)
	if err != nil {
		return fmt.Errorf("failed to write promotion ID: %w", err)
	}

	// Separator
	sep2, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Tombstone (1 byte) - 0x00 for active
	tombstone := []byte{0x00}

	// Record separator
	recSep, err := utils.WriteVariable(utils.RecordSeparator)
	if err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, orderIDBytes...)
	entry = append(entry, sep1...)
	entry = append(entry, promotionIDBytes...)
	entry = append(entry, sep2...)
	entry = append(entry, tombstone...)
	entry = append(entry, recSep...)

	// Write entry to file
	_, err = file.Write(entry)
	if err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	// Update header with new entity count
	err = utils.UpdateHeader(file, entitiesCount+1, tombstoneCount, nextId)
	if err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	// Sync to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync to disk: %w", err)
	}

	return nil
}

// existsUnlocked checks if a relationship already exists (must be called with lock held)
func (dao *OrderPromotionDAO) existsUnlocked(orderID, promotionID uint64) (bool, error) {
	// Split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Check each entry for matching composite key
	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < utils.IDSize*2+utils.TombstoneSize+2 {
			continue
		}

		offset := 0

		// Read orderID
		entryOrderID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip separator
		offset += 1

		// Read promotionID
		entryPromotionID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip separator
		offset += 1

		// Read tombstone
		tombstone := entryData[offset]

		// Check if this is an active entry with matching composite key
		if tombstone == 0x00 && entryOrderID == orderID && entryPromotionID == promotionID {
			return true, nil
		}
	}

	return false, nil
}

// GetByOrderID retrieves all promotions applied to an order
func (dao *OrderPromotionDAO) GetByOrderID(orderID uint64) ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Parse each entry and filter by orderID
	// New format: [orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)]
	result := make([]*OrderPromotion, 0)

	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < utils.IDSize*2+utils.TombstoneSize+2 {
			continue
		}

		offset := 0

		// Read orderID
		entryOrderID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read tombstone
		tombstone := entryData[offset]

		// Skip deleted entries
		if tombstone != 0x00 {
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

	// Split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Parse each entry and filter by promotionID
	// New format: [orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)]
	result := make([]*OrderPromotion, 0)

	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < utils.IDSize*2+utils.TombstoneSize+2 {
			continue
		}

		offset := 0

		// Read orderID
		entryOrderID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read tombstone
		tombstone := entryData[offset]

		// Skip deleted entries
		if tombstone != 0x00 {
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
// Finds entry by composite key (orderID, promotionID)
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

	// Split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to split file into entries: %w", err)
	}

	// Find the entry with matching composite key (orderID, promotionID)
	// New format: [orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)]
	for _, entry := range entries {
		entryData := entry.Data
		if len(entryData) < utils.IDSize*2+utils.TombstoneSize+2 {
			continue
		}

		offset := 0

		// Read orderID
		entryOrderID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Read promotionID
		entryPromotionID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}
		offset = newOffset

		// Skip unit separator
		offset += 1

		// Check tombstone
		tombstone := entryData[offset]

		if tombstone != 0x00 {
			continue // Already deleted
		}

		// Check if this is the entry we want to delete (by composite key)
		if entryOrderID == orderID && entryPromotionID == promotionID {
			// Calculate tombstone position
			// Position = entryStart + orderID(2) + sep(1) + promotionID(2) + sep(1)
			tombstonePos := entry.Position + int64(utils.IDSize) + 1 + int64(utils.IDSize) + 1

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

			// Update header (decrement entities, increment tombstones)
			err = utils.UpdateHeader(file, entitiesCount-1, tombstoneCount+1, nextId)
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
