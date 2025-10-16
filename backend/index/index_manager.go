package index

import (
	"BinaryCRUD/backend/index/b_tree"
	"BinaryCRUD/backend/serialization"
	"bufio"
	"fmt"
	"io"
	"os"
)

// IndexManager manages the B+ tree index for a binary data file
type IndexManager struct {
	tree         *b_tree.BPlusTree
	dataFilename string
	indexFilename string
	order        int
}

// NewIndexManager creates a new index manager
// dataFilename: path to the binary data file (e.g., "data/items.bin")
// order: B+ tree order (max keys per node, default 4)
func NewIndexManager(dataFilename string, order int) *IndexManager {
	if order < 3 {
		order = 4 // default order
	}

	indexFilename := dataFilename + ".idx"

	return &IndexManager{
		tree:          nil, // initialized lazily
		dataFilename:  dataFilename,
		indexFilename: indexFilename,
		order:         order,
	}
}

// Initialize loads or builds the index
// Should be called once at startup
func (m *IndexManager) Initialize() error {
	fmt.Printf("[INDEX] Initializing index for %s\n", m.dataFilename)

	// Try to load existing index file
	if b_tree.IndexFileExists(m.indexFilename) {
		fmt.Printf("[INDEX] Loading existing index from %s\n", m.indexFilename)
		tree, err := b_tree.LoadFromFile(m.indexFilename)
		if err != nil {
			fmt.Printf("[INDEX] Failed to load index: %v, rebuilding...\n", err)
			return m.RebuildIndex()
		}
		m.tree = tree
		fmt.Printf("[INDEX] Index loaded: %d entries\n", m.tree.Count())
		return nil
	}

	// No index file exists, build from data file
	fmt.Printf("[INDEX] No index file found, building from data file...\n")
	return m.RebuildIndex()
}

// RebuildIndex scans the data file and rebuilds the index from scratch
func (m *IndexManager) RebuildIndex() error {
	fmt.Printf("[INDEX] Rebuilding index from %s\n", m.dataFilename)

	// Create new tree
	m.tree = b_tree.NewBPlusTree(m.order)

	// Check if data file exists
	if _, err := os.Stat(m.dataFilename); os.IsNotExist(err) {
		fmt.Printf("[INDEX] Data file doesn't exist yet, creating empty index\n")
		return m.Save()
	}

	// Open data file
	file, err := os.Open(m.dataFilename)
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read header to get record count
	count, err := serialization.ReadHeader(reader)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	fmt.Printf("[INDEX] Scanning %d records...\n", count)

	// Track file position (starts after header)
	format := serialization.GetFormat()
	currentOffset := int64(format.HeaderSize())

	// Read each record and build index
	for i := uint32(0); i < count; i++ {
		// Record the offset BEFORE reading the record
		recordOffset := currentOffset

		// Read the record
		item, err := serialization.ReadRecord(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read record %d: %w", i, err)
		}

		// Calculate size of this record
		recordSize := format.CalculateRecordSize(len(item.Name))
		currentOffset += int64(recordSize)

		// Insert into index using the stored ID from the record (ID → FileOffset)
		m.tree.Insert(item.RecordID, recordOffset)

		if i < 5 || i >= count-5 {
			fmt.Printf("[INDEX]   Record %d (ID=%d): offset=%d, name=%s\n", i, item.RecordID, recordOffset, item.Name)
		} else if i == 5 {
			fmt.Printf("[INDEX]   ... (%d more records) ...\n", count-10)
		}
	}

	fmt.Printf("[INDEX] Index rebuilt: %d entries\n", m.tree.Count())

	// Save to disk
	return m.Save()
}

// Insert adds a new entry to the index
// recordID: the position of the record in the file (0-based)
// offset: the byte offset where the record starts in the file
func (m *IndexManager) Insert(recordID uint32, offset int64) error {
	if m.tree == nil {
		return fmt.Errorf("index not initialized")
	}

	fmt.Printf("[INDEX] Inserting: recordID=%d, offset=%d\n", recordID, offset)
	m.tree.Insert(recordID, offset)
	return nil
}

// GetOffset looks up the file offset for a given record ID
// Returns: offset, found
func (m *IndexManager) GetOffset(recordID uint32) (int64, bool) {
	if m.tree == nil {
		return 0, false
	}

	// Check if index file actually exists on disk
	// If it was deleted, don't use the in-memory tree
	if !b_tree.IndexFileExists(m.indexFilename) {
		fmt.Printf("[INDEX] Index file no longer exists on disk\n")
		return 0, false
	}

	return m.tree.Search(recordID)
}

// Save persists the index to disk
func (m *IndexManager) Save() error {
	if m.tree == nil {
		return fmt.Errorf("index not initialized")
	}

	return m.tree.SaveToFile(m.indexFilename)
}

// GetRecordByID reads a specific record from the data file using the index
// Returns the item at the given record ID
// If the index lookup fails, it falls back to sequential search
func (m *IndexManager) GetRecordByID(recordID uint32) (*serialization.Item, error) {
	// Try to lookup offset in index first
	offset, found := m.GetOffset(recordID)

	if found {
		// Index lookup succeeded - use direct file access
		fmt.Printf("[INDEX] Found recordID=%d at offset=%d (via index)\n", recordID, offset)

		// Open data file
		file, err := os.Open(m.dataFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to open data file: %w", err)
		}
		defer file.Close()

		// Seek to offset
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek to offset %d: %w", offset, err)
		}

		// Read the record
		reader := bufio.NewReader(file)
		item, err := serialization.ReadRecord(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read record at offset %d: %w", offset, err)
		}

		return item, nil
	}

	// Index lookup failed - fall back to sequential search
	fmt.Printf("[INDEX] RecordID=%d not found in index, falling back to sequential search\n", recordID)
	return m.sequentialSearchByID(recordID)
}

// sequentialSearchByID performs a sequential search through the data file to find a record by ID
// This is used as a fallback when the index is not available or doesn't contain the record
func (m *IndexManager) sequentialSearchByID(recordID uint32) (*serialization.Item, error) {
	// Check if data file exists
	if _, err := os.Stat(m.dataFilename); os.IsNotExist(err) {
		return nil, fmt.Errorf("data file does not exist")
	}

	// Open data file
	file, err := os.Open(m.dataFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read header to get record count
	count, err := serialization.ReadHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	fmt.Printf("[INDEX] Sequentially searching for recordID=%d among %d records\n", recordID, count)

	// Read records sequentially until we find the target ID
	for i := uint32(0); i < count; i++ {
		item, err := serialization.ReadRecord(reader)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("unexpected EOF while searching for record ID %d", recordID)
			}
			return nil, fmt.Errorf("failed to read record at position %d: %w", i, err)
		}

		// Found the target record by comparing stored IDs
		if item.RecordID == recordID {
			fmt.Printf("[INDEX] Found recordID=%d via sequential search\n", recordID)
			return item, nil
		}
	}

	// Record ID not found in the file
	return nil, fmt.Errorf("record ID %d not found after scanning %d records", recordID, count)
}

// PrintTree prints the B+ tree structure (for debugging)
func (m *IndexManager) PrintTree() {
	if m.tree != nil {
		m.tree.PrintTree()
	}
}

// GetCurrentOffset calculates the file offset where the next record should be written
// This reads the data file to find the current end position
func (m *IndexManager) GetCurrentOffset() (int64, error) {
	file, err := os.Open(m.dataFilename)
	if err != nil {
		// If file doesn't exist, offset is 0 (will be header size after init)
		if os.IsNotExist(err) {
			format := serialization.GetFormat()
			return int64(format.HeaderSize()), nil
		}
		return 0, fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat data file: %w", err)
	}

	return fileInfo.Size(), nil
}

// UpdateHeader updates the record count in the data file header
// Called after appending new records
// Note: This function is currently unused as AppendEntry handles header updates directly
func (m *IndexManager) UpdateHeader(count uint32) error {
	file, err := os.OpenFile(m.dataFilename, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open data file for header update: %w", err)
	}
	defer file.Close()

	// Seek to beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to header: %w", err)
	}

	// Write updated header using centralized writer
	writer := bufio.NewWriter(file)
	if err := serialization.WriteHeader(writer, count); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush header: %w", err)
	}

	return nil
}
