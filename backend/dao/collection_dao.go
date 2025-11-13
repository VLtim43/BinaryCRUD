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
// Binary format: [ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)][itemID2(2)]...[0x1E]
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

	// Build entry without ID: [0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)]...

	// Separator before nameSize
	sep1, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return 0, fmt.Errorf("failed to write separator: %w", err)
	}

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

	// Separator before total price
	sep2, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return 0, fmt.Errorf("failed to write separator: %w", err)
	}

	// Total price (4 bytes)
	totalPriceBytes, err := utils.WriteFixedNumber(4, totalPrice)
	if err != nil {
		return 0, fmt.Errorf("failed to write total price: %w", err)
	}

	// Separator before item count
	sep3, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return 0, fmt.Errorf("failed to write separator: %w", err)
	}

	// Item count (4 bytes)
	itemCount := uint64(len(itemIDs))
	itemCountBytes, err := utils.WriteFixedNumber(4, itemCount)
	if err != nil {
		return 0, fmt.Errorf("failed to write item count: %w", err)
	}

	// Separator before item IDs
	sep4, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return 0, fmt.Errorf("failed to write separator: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, sep1...)
	entry = append(entry, nameSizeBytes...)
	entry = append(entry, nameBytes...)
	entry = append(entry, sep2...)
	entry = append(entry, totalPriceBytes...)
	entry = append(entry, sep3...)
	entry = append(entry, itemCountBytes...)
	entry = append(entry, sep4...)

	// Add all item IDs (2 bytes each)
	for _, itemID := range itemIDs {
		itemIDBytes, err := utils.WriteFixedNumber(utils.IDSize, itemID)
		if err != nil {
			return 0, fmt.Errorf("failed to write item ID: %w", err)
		}
		entry = append(entry, itemIDBytes...)
	}

	// Read header to get the next ID
	_, _, nextId, err := utils.ReadHeader(file)
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
		_, err = file.Seek(offset, 0)
		if err == nil {
			// Read until record separator
			buffer := make([]byte, 8192) // Larger buffer for collections with many items
			n, err := file.Read(buffer)
			if err == nil {
				// Find record separator
				recordSep := []byte(utils.RecordSeparator)[0]
				sepPos := -1
				for i := 0; i < n; i++ {
					if buffer[i] == recordSep {
						sepPos = i
						break
					}
				}

				if sepPos != -1 {
					entryData = buffer[:sepPos]
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

// GetAll retrieves all non-deleted collections using sequential scan
func (dao *CollectionDAO) GetAll() ([]*Collection, error) {
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
	result := make([]*Collection, 0)

	for _, entry := range entries {
		// Parse the entry using utility function
		collection, err := utils.ParseCollectionEntry(entry.Data)
		if err != nil {
			// Skip malformed entries
			continue
		}

		// Skip deleted entries
		if collection.Tombstone != 0x00 {
			continue
		}

		result = append(result, &Collection{
			ID:          collection.ID,
			OwnerOrName: collection.OwnerOrName,
			TotalPrice:  collection.TotalPrice,
			ItemCount:   collection.ItemCount,
			ItemIDs:     collection.ItemIDs,
		})
	}

	return result, nil
}
