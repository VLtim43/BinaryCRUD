package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
	"fmt"
	"io"
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
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Open file for reading
	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	var entryData []byte

	// Try B+ tree index first
	offset, found := dao.tree.Search(id)
	if found {
		// Use index for fast lookup
		// offset points to the start of the record (at the length prefix)
		_, err = file.Seek(offset, 0)
		if err == nil {
			// Read record length (2 bytes)
			lengthBytes := make([]byte, utils.RecordLengthSize)
			n, err := file.Read(lengthBytes)
			if err == nil && n == utils.RecordLengthSize {
				recordLength, _, err := utils.ReadFixedNumber(utils.RecordLengthSize, lengthBytes, 0)
				if err == nil {
					// Read the record data (after length prefix)
					entryData = make([]byte, recordLength)
					n, err := file.Read(entryData)
					if err != nil || n != int(recordLength) {
						entryData = nil
					}
				}
			}
		}
	}

	// If index lookup failed or didn't find data, fall back to sequential scan
	if entryData == nil {
		entryData, err = utils.FindByIDSequential(file, id)
		if err != nil {
			return nil, err
		}
	}

	// Parse the entry using utility function
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
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Create index delete function
	indexDeleteFunc := func(id uint64) error {
		// Remove from B+ tree index
		err := dao.tree.Delete(id)
		if err != nil {
			return err
		}

		// Save updated index to disk
		return dao.tree.Save(dao.indexPath)
	}

	// Use the generic soft delete utility
	return utils.SoftDeleteByID(dao.filePath, id, &dao.mu, indexDeleteFunc)
}

// GetAll retrieves all collections from the database, including deleted ones
func (dao *CollectionDAO) GetAll() ([]*Collection, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Open file for reading
	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	// Seek past header
	_, err = file.Seek(int64(utils.HeaderSize), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek past header: %w", err)
	}

	// Read all file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result := make([]*Collection, 0)

	// Parse all records sequentially
	offset := 0
	for offset < len(fileData) {
		// Check if we have enough bytes for the length field
		if offset+utils.RecordLengthSize > len(fileData) {
			break
		}

		// Read the record length
		recordLength, lengthEnd, err := utils.ReadFixedNumber(utils.RecordLengthSize, fileData, offset)
		if err != nil {
			break
		}

		// Check if we have enough bytes for the complete record
		if lengthEnd+int(recordLength) > len(fileData) {
			break
		}

		// Extract the record data (without length prefix)
		entryData := fileData[lengthEnd : lengthEnd+int(recordLength)]

		// Parse the entry
		collection, err := utils.ParseCollectionEntry(entryData)
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

		// Move to next record
		offset = lengthEnd + int(recordLength)
	}

	return result, nil
}
