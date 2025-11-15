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
// Binary format with composite primary key: [recordLength(2)][orderID(2)][promotionID(2)][tombstone(1)]
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
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open order_promotion file: %w", err)
	}
	defer file.Close()

	// Build entry data: [orderID(2)][promotionID(2)][tombstone(1)]

	// Order ID (2 bytes)
	orderIDBytes, err := utils.WriteFixedNumber(utils.IDSize, orderID)
	if err != nil {
		return fmt.Errorf("failed to write order ID: %w", err)
	}

	// Promotion ID (2 bytes)
	promotionIDBytes, err := utils.WriteFixedNumber(utils.IDSize, promotionID)
	if err != nil {
		return fmt.Errorf("failed to write promotion ID: %w", err)
	}

	// Tombstone (1 byte) - 0x00 for active
	tombstone := []byte{0x00}

	// Combine all fields: [orderID][promotionID][tombstone]
	entryData := make([]byte, 0, utils.IDSize+utils.IDSize+utils.TombstoneSize)
	entryData = append(entryData, orderIDBytes...)
	entryData = append(entryData, promotionIDBytes...)
	entryData = append(entryData, tombstone...)

	// Use the manual append utility to write the entry with proper formatting and header updates
	err = utils.AppendEntryManual(file, entryData)
	if err != nil {
		return fmt.Errorf("failed to append entry: %w", err)
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

	// Check each entry for matching composite key using utility parser
	for _, entry := range entries {
		// Parse the entry using utility function
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			// Skip malformed entries
			continue
		}

		// Check if this is an active entry with matching composite key
		if orderPromo.Tombstone == 0x00 && orderPromo.OrderID == orderID && orderPromo.PromotionID == promotionID {
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

	// Parse each entry and filter by orderID using utility function
	result := make([]*OrderPromotion, 0)

	for _, entry := range entries {
		// Parse the entry using utility function
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			// Skip malformed entries
			continue
		}

		// Skip deleted entries
		if orderPromo.Tombstone != 0x00 {
			continue
		}

		// Filter by orderID
		if orderPromo.OrderID == orderID {
			result = append(result, &OrderPromotion{
				OrderID:     orderPromo.OrderID,
				PromotionID: orderPromo.PromotionID,
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

	// Parse each entry and filter by promotionID using utility function
	result := make([]*OrderPromotion, 0)

	for _, entry := range entries {
		// Parse the entry using utility function
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			// Skip malformed entries
			continue
		}

		// Skip deleted entries
		if orderPromo.Tombstone != 0x00 {
			continue
		}

		// Filter by promotionID
		if orderPromo.PromotionID == promotionID {
			result = append(result, &OrderPromotion{
				OrderID:     orderPromo.OrderID,
				PromotionID: orderPromo.PromotionID,
			})
		}
	}

	return result, nil
}

// GetAll retrieves all non-deleted order-promotion relationships
func (dao *OrderPromotionDAO) GetAll() ([]*OrderPromotion, error) {
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

	// Parse each entry using utility function
	result := make([]*OrderPromotion, 0)

	for _, entry := range entries {
		// Parse the entry using utility function
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			// Skip malformed entries
			continue
		}

		// Skip deleted entries
		if orderPromo.Tombstone != 0x00 {
			continue
		}

		result = append(result, &OrderPromotion{
			OrderID:     orderPromo.OrderID,
			PromotionID: orderPromo.PromotionID,
		})
	}

	return result, nil
}

// Delete removes an order-promotion relationship by marking it as deleted
// Finds entry by composite key (orderID, promotionID)
func (dao *OrderPromotionDAO) Delete(orderID, promotionID uint64) error {
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Use the generic soft delete utility for composite keys
	return utils.SoftDeleteByCompositeKey(dao.filePath, orderID, promotionID, &dao.mu)
}
