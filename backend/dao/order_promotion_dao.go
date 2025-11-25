package dao

import (
	"BinaryCRUD/backend/index"
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
	filePath  string
	indexPath string
	hashIndex *index.ExtensibleHash
	mu        sync.Mutex
}

// NewOrderPromotionDAO creates a DAO for order_promotions.bin
func NewOrderPromotionDAO(filePath string) *OrderPromotionDAO {
	// Use the utility function that handles rebuild on corruption
	indexPath, hashIndex := utils.InitializeOrderPromotionIndex(filePath, 4)

	return &OrderPromotionDAO{
		filePath:  filePath,
		indexPath: indexPath,
		hashIndex: hashIndex,
	}
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

	// Check for duplicate using hash index (fast O(1) lookup)
	_, exists := dao.hashIndex.Search(orderID, promotionID)
	if exists {
		return fmt.Errorf("order-promotion relationship already exists (orderID=%d, promotionID=%d)", orderID, promotionID)
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open order_promotion file: %w", err)
	}
	defer file.Close()

	// Get current file offset before writing (for index)
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	entryOffset := fileInfo.Size()

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

	// Add to hash index
	err = dao.hashIndex.Insert(orderID, promotionID, entryOffset)
	if err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	// Persist index
	err = dao.hashIndex.Save(dao.indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// GetByOrderID retrieves all promotions applied to an order
func (dao *OrderPromotionDAO) GetByOrderID(orderID uint64) ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Use hash index for fast lookup
	entries := dao.hashIndex.GetByOrderID(orderID)

	result := make([]*OrderPromotion, len(entries))
	for i, entry := range entries {
		result[i] = &OrderPromotion{
			OrderID:     entry.OrderID,
			PromotionID: entry.PromotionID,
		}
	}

	return result, nil
}

// GetByPromotionID retrieves all orders that have a specific promotion applied
func (dao *OrderPromotionDAO) GetByPromotionID(promotionID uint64) ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Use hash index for fast lookup
	entries := dao.hashIndex.GetByPromotionID(promotionID)

	result := make([]*OrderPromotion, len(entries))
	for i, entry := range entries {
		result[i] = &OrderPromotion{
			OrderID:     entry.OrderID,
			PromotionID: entry.PromotionID,
		}
	}

	return result, nil
}

// GetAll retrieves all non-deleted order-promotion relationships
func (dao *OrderPromotionDAO) GetAll() ([]*OrderPromotion, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Use hash index for fast retrieval
	entries := dao.hashIndex.GetAll()

	result := make([]*OrderPromotion, len(entries))
	for i, entry := range entries {
		result[i] = &OrderPromotion{
			OrderID:     entry.OrderID,
			PromotionID: entry.PromotionID,
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

	// Remove from hash index first
	err := dao.hashIndex.Delete(orderID, promotionID)
	if err != nil {
		return fmt.Errorf("key not found: orderID=%d, promotionID=%d", orderID, promotionID)
	}

	// Save updated index
	err = dao.hashIndex.Save(dao.indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Use the generic soft delete utility for composite keys (without mutex since we already hold it)
	return utils.SoftDeleteByCompositeKey(dao.filePath, orderID, promotionID, nil)
}

// GetHashIndex returns the hash index for debugging/inspection
func (dao *OrderPromotionDAO) GetHashIndex() *index.ExtensibleHash {
	return dao.hashIndex
}
