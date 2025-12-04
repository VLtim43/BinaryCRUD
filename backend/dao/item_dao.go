package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/search"
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"strings"
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
	entry := utils.CombineBytes(nameSizeBytes, nameBytes, priceBytes)

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

// Read retrieves an item by ID using the B+ tree index with automatic fallback to sequential scan
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) Read(id uint64) (uint64, string, uint64, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Open file for reading (don't create if it doesn't exist)
	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", 0, fmt.Errorf("failed to open item file: file does not exist")
		}
		return 0, "", 0, fmt.Errorf("failed to open item file: %w", err)
	}
	defer file.Close()

	var entryData []byte

	// Try B+ tree index first
	if offset, found := dao.tree.Search(id); found {
		var readErr error
		entryData, readErr = utils.ReadEntryAtOffset(file, offset)
		if readErr != nil {
			// Index may be stale, log and fall back to sequential scan
			entryData = nil
		}
	}

	// If index lookup failed or returned no data, fall back to sequential scan
	if entryData == nil {
		entryData, err = utils.FindByIDSequential(file, id)
		if err != nil {
			return 0, "", 0, fmt.Errorf("item not found: %w", err)
		}
	}

	// Parse the entry
	item, err := utils.ParseItemEntry(entryData)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to parse item entry: %w", err)
	}

	// Check if item is deleted
	if item.Tombstone != 0x00 {
		return 0, "", 0, fmt.Errorf("deleted item id %d", item.ID)
	}

	return item.ID, item.Name, item.Price, nil
}

// Delete marks an item as deleted by flipping its tombstone bit
// This is a logical deletion - the data remains in the file but is marked as deleted
func (dao *ItemDAO) Delete(id uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	return utils.DeleteFromBTreeIndex(dao.tree, dao.indexPath, dao.filePath, id, "item")
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
	IsDeleted    bool
}

// GetAll retrieves all items from the database, including deleted ones
func (dao *ItemDAO) GetAll() ([]Item, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(dao.filePath); os.IsNotExist(err) {
		return []Item{}, nil
	}

	// Use utility to split file into entries
	entries, err := utils.SplitFileIntoEntries(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read items: %w", err)
	}

	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		item, err := utils.ParseItemEntry(entry.Data)
		if err == nil {
			items = append(items, Item{
				ID:           item.ID,
				Name:         item.Name,
				PriceInCents: item.Price,
				IsDeleted:    item.Tombstone != 0x00,
			})
		}
	}

	return items, nil
}

// SearchAlgorithm represents the pattern matching algorithm to use
type SearchAlgorithm string

const (
	AlgorithmKMP        SearchAlgorithm = "kmp"
	AlgorithmBoyerMoore SearchAlgorithm = "bm"
)

// SearchByName finds items whose name contains the given pattern.
// Returns only non-deleted items. Case-insensitive search.
// algorithm: "kmp" for Knuth-Morris-Pratt, "bm" for Boyer-Moore
func (dao *ItemDAO) SearchByName(pattern string, algorithm SearchAlgorithm) ([]Item, error) {
	items, err := dao.GetAll()
	if err != nil {
		return nil, err
	}

	if pattern == "" {
		return []Item{}, nil
	}

	// Case-insensitive: lowercase the pattern
	lowerPattern := strings.ToLower(pattern)

	// Create matcher based on algorithm choice
	var matcher interface {
		ContainsString(text string) bool
	}

	switch algorithm {
	case AlgorithmBoyerMoore:
		matcher = search.NewBoyerMooreString(lowerPattern)
	case AlgorithmKMP:
		fallthrough
	default:
		matcher = search.NewKMPString(lowerPattern)
	}

	var results []Item
	for _, item := range items {
		if item.IsDeleted {
			continue
		}
		// Case-insensitive: lowercase the name before matching
		if matcher.ContainsString(strings.ToLower(item.Name)) {
			results = append(results, item)
		}
	}

	return results, nil
}
