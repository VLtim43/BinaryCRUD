package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"sync"
)

// ItemDAO manages the items binary file
type ItemDAO struct {
	filePath  string
	indexPath string
	mu        sync.Mutex    // Protects concurrent writes to the binary file
	tree      *index.BTree  // B+ tree index for fast lookups
}

// NewItemDAO creates a new ItemDAO instance
func NewItemDAO(filePath string) *ItemDAO {
	indexPath, tree := utils.InitializeDAOIndex(filePath)

	return &ItemDAO{
		filePath:  filePath,
		indexPath: indexPath,
		tree:      tree,
	}
}

// ensureFileExists creates the file with empty header if it doesn't exist
func (dao *ItemDAO) ensureFileExists() error {
	return utils.EnsureFileExists(dao.filePath)
}

// Write adds an item to the binary file and returns the assigned ID
// Complete record structure: [recordLength(2)][ID(2)][tombstone(1)][nameLength(2)][name...][price(4)]
// ID, tombstone, and record length are auto-assigned by AppendEntry (tombstone is 0x00 for active records)
func (dao *ItemDAO) Write(name string, priceInCents uint64) (uint64, error) {
	// Lock to prevent concurrent writes
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return 0, err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open item file: %w", err)
	}
	defer file.Close()

	// Calculate name size
	nameSize := len(name)

	// Build entry without ID and tombstone: [nameLength(2)][name...][price(4)]
	// ID and tombstone will be added by AppendEntry

	// Name size (2 bytes - supports names up to 65535 chars)
	nameSizeBytes, err := utils.WriteFixedNumber(2, uint64(nameSize))
	if err != nil {
		return 0, fmt.Errorf("failed to write name size: %w", err)
	}

	// Name (variable length)
	nameBytes, err := utils.WriteVariable(name)
	if err != nil {
		return 0, fmt.Errorf("failed to write name: %w", err)
	}

	// Price (4 bytes - supports prices up to 4,294,967,295 cents)
	priceBytes, err := utils.WriteFixedNumber(4, priceInCents)
	if err != nil {
		return 0, fmt.Errorf("failed to write price: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, nameSizeBytes...)
	entry = append(entry, nameBytes...)
	entry = append(entry, priceBytes...)

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

	// Append the entry (ID auto-assigned and record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return 0, fmt.Errorf("failed to append item: %w", err)
	}

	// Add to index: ID -> file offset
	dao.tree.Insert(uint64(nextId), appendPos)

	// Save index to disk
	err = dao.tree.Save(dao.indexPath)
	if err != nil {
		return 0, fmt.Errorf("failed to save index: %w", err)
	}

	return uint64(nextId), nil
}

// Read retrieves an item by ID (uses index with automatic fallback)
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) Read(id uint64) (uint64, string, uint64, error) {
	return dao.readWithIndex(id)
}

// ReadWithIndex retrieves an item by ID using the B+ tree index with automatic fallback to sequential scan
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) ReadWithIndex(id uint64, useIndex bool) (uint64, string, uint64, error) {
	return dao.readWithIndex(id)
}

// readWithIndex is the internal implementation that always tries index first, then falls back to sequential
func (dao *ItemDAO) readWithIndex(id uint64) (uint64, string, uint64, error) {
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return 0, "", 0, err
	}

	// Open file for reading
	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to open item file: %w", err)
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
		// Use sequential finder to locate the entry
		entryData, err = utils.FindByIDSequential(file, id)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to find item: %w", err)
		}
	}

	// Parse the entry using utility function
	item, err := utils.ParseItemEntry(entryData)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to parse item entry: %w", err)
	}

	// Check if item is deleted
	if item.Tombstone != 0x00 {
		return 0, "", 0, fmt.Errorf("deleted file id %d", item.ID)
	}

	return item.ID, item.Name, item.Price, nil
}

// Delete marks an item as deleted by flipping its tombstone bit
// This is a logical deletion - the data remains in the file but is marked as deleted
func (dao *ItemDAO) Delete(id uint64) error {
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Create index delete function
	indexDeleteFunc := func(id uint64) error {
		// Remove from index
		err := dao.tree.Delete(id)
		if err != nil {
			return err
		}

		// Save updated index
		return dao.tree.Save(dao.indexPath)
	}

	// Use the generic soft delete utility
	return utils.SoftDeleteByID(dao.filePath, id, &dao.mu, indexDeleteFunc)
}

// GetIndexTree returns the B+ tree for debugging purposes
func (dao *ItemDAO) GetIndexTree() *index.BTree {
	return dao.tree
}

// Item represents an item record
type Item struct {
	ID           uint64
	Name         string
	PriceInCents uint64
}

// GetAll retrieves all non-deleted items from the database
func (dao *ItemDAO) GetAll() ([]Item, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Get all entries from the index
	allEntries := dao.tree.GetAll()

	// Collect IDs and sort them for consistent ordering
	ids := make([]uint64, 0, len(allEntries))
	for id := range allEntries {
		ids = append(ids, id)
	}

	// Sort IDs for consistent, predictable ordering
	for i := 0; i < len(ids)-1; i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}

	items := make([]Item, 0, len(ids))

	// Read each item by ID in sorted order
	for _, id := range ids {
		itemID, name, priceInCents, err := dao.readWithIndex(id)
		if err != nil {
			// Skip deleted or errored items
			continue
		}

		items = append(items, Item{
			ID:           itemID,
			Name:         name,
			PriceInCents: priceInCents,
		})
	}

	return items, nil
}
