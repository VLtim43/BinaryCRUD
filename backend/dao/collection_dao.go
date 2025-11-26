package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"sync"
)

// Collection represents either an Order or Promotion
// - For Orders: OwnerOrName is the customer name
// - For Promotions: OwnerOrName is the promotion name
type Collection struct {
	ID          uint64
	OwnerOrName string
	TotalPrice  uint64
	ItemCount   uint64
	ItemIDs     []uint64
	IsDeleted   bool
}

type CollectionDAO struct {
	filePath  string
	indexPath string
	mu        sync.Mutex
	tree      *index.BTree // B+ tree index for fast lookups
}

// ensureFileExists creates the file with empty header if it doesn't exist
func (dao *CollectionDAO) ensureFileExists() error {
	return utils.EnsureFileExists(dao.filePath)
}

// Write creates a new collection entry and returns the assigned ID
// Complete record format: [recordLength(2)][ID(2)][tombstone(1)][nameLength(2)][name...][totalPrice(4)][itemCount(4)][itemIDs...]
func (dao *CollectionDAO) Write(ownerOrName string, totalPrice uint64, itemIDs []uint64) (uint64, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return 0, err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	// Build entry without ID and tombstone: [nameLength(2)][name...][totalPrice(4)][itemCount(4)][itemIDs...]
	// ID, tombstone, and record length will be added by AppendEntry

	// Name size (2 bytes)
	nameSize := len(ownerOrName)
	nameSizeBytes, err := utils.WriteFixedNumber(2, uint64(nameSize))
	if err != nil {
		return 0, fmt.Errorf("failed to write name size: %w", err)
	}

	// Name (variable length)
	nameBytes, err := utils.WriteVariable(ownerOrName)
	if err != nil {
		return 0, fmt.Errorf("failed to write name: %w", err)
	}

	// Total price (4 bytes)
	totalPriceBytes, err := utils.WriteFixedNumber(4, totalPrice)
	if err != nil {
		return 0, fmt.Errorf("failed to write total price: %w", err)
	}

	// Item count (4 bytes)
	itemCount := uint64(len(itemIDs))
	itemCountBytes, err := utils.WriteFixedNumber(4, itemCount)
	if err != nil {
		return 0, fmt.Errorf("failed to write item count: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, nameSizeBytes...)
	entry = append(entry, nameBytes...)
	entry = append(entry, totalPriceBytes...)
	entry = append(entry, itemCountBytes...)

	// Add all item IDs (2 bytes each)
	for _, itemID := range itemIDs {
		itemIDBytes, err := utils.WriteFixedNumber(utils.IDSize, itemID)
		if err != nil {
			return 0, fmt.Errorf("failed to write item ID: %w", err)
		}
		entry = append(entry, itemIDBytes...)
	}

	// Read header to get the next ID
	_, _, _, nextId, err := utils.ReadHeader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	// Seek back to end
	_, err = file.Seek(0, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to seek to end: %w", err)
	}

	// Get actual append position
	appendPos, err := file.Seek(0, 1)
	if err != nil {
		return 0, fmt.Errorf("failed to get append position: %w", err)
	}

	// Append the entry (ID and tombstone auto-assigned, record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return 0, fmt.Errorf("failed to append collection: %w", err)
	}

	// Add to B+ tree index: ID -> file offset
	dao.tree.Insert(uint64(nextId), appendPos)

	// Save index to disk
	err = dao.tree.Save(dao.indexPath)
	if err != nil {
		return 0, fmt.Errorf("failed to save index: %w", err)
	}

	return uint64(nextId), nil
}

// Read retrieves a collection by ID using B+ tree index with automatic fallback to sequential scan
func (dao *CollectionDAO) Read(id uint64) (*Collection, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	return dao.readUnlocked(id)
}

// readUnlocked is the internal implementation (must be called with lock held)
func (dao *CollectionDAO) readUnlocked(id uint64) (*Collection, error) {
	// Open file for reading (don't create if it doesn't exist)
	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("collection file does not exist")
		}
		return nil, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	var entryData []byte

	// Try B+ tree index first
	if offset, found := dao.tree.Search(id); found {
		entryData, _ = utils.ReadEntryAtOffset(file, offset)
	}

	// If index lookup failed, fall back to sequential scan
	if entryData == nil {
		entryData, err = utils.FindByIDSequential(file, id)
		if err != nil {
			return nil, err
		}
	}

	// Parse the entry
	collection, err := utils.ParseCollectionEntry(entryData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection entry: %w", err)
	}

	// Check if deleted
	if collection.Tombstone != 0x00 {
		return nil, fmt.Errorf("collection with ID %d is deleted", collection.ID)
	}

	return &Collection{
		ID:          collection.ID,
		OwnerOrName: collection.OwnerOrName,
		TotalPrice:  collection.TotalPrice,
		ItemCount:   collection.ItemCount,
		ItemIDs:     collection.ItemIDs,
	}, nil
}

// Delete marks a collection as deleted by flipping the tombstone bit
func (dao *CollectionDAO) Delete(id uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Remove from index first
	err := dao.tree.Delete(id)
	if err != nil {
		return fmt.Errorf("collection not found in index: %w", err)
	}

	// Save updated index
	err = dao.tree.Save(dao.indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Use the generic soft delete utility (nil mutex since we already hold the lock)
	return utils.SoftDeleteByID(dao.filePath, id, nil, nil)
}

// GetAll retrieves all collections from the database, including deleted ones
func (dao *CollectionDAO) GetAll() ([]*Collection, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(dao.filePath); os.IsNotExist(err) {
		return []*Collection{}, nil
	}

	// Use utility to split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read collections: %w", err)
	}

	result := make([]*Collection, 0, len(entries))
	for _, entry := range entries {
		collection, err := utils.ParseCollectionEntry(entry.Data)
		if err == nil {
			result = append(result, &Collection{
				ID:          collection.ID,
				OwnerOrName: collection.OwnerOrName,
				TotalPrice:  collection.TotalPrice,
				ItemCount:   collection.ItemCount,
				ItemIDs:     collection.ItemIDs,
				IsDeleted:   collection.Tombstone != 0x00,
			})
		}
	}

	return result, nil
}
