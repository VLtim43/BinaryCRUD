package dao

// Binary structure of item.bin:
//
// HEADER:
//   [Count:4 bytes][0x1E]
//
// ITEM RECORD:
//   [RecordID:4 bytes][0x1F][Tombstone:1 byte][0x1F][NameLength:4 bytes][0x1F][Name:N bytes][0x1F][Timestamp:8 bytes][0x1E]
//
// Notes:
//   - All multi-byte integers are little-endian
//   - Tombstone: 0=active, 1=deleted
//   - RecordID is assigned from Count before incrementing
//   - 0x1F = Unit Separator, 0x1E = Record Separator

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/persistence"
	"fmt"
	"os"
)

type ItemDAO struct {
	filename     string
	indexManager *index.IndexManager
}

func NewItemDAO(filename string) *ItemDAO {
	// Create index manager with B+ tree order of 4
	indexManager := index.NewIndexManager(filename, 4)

	// Initialize the index (load or rebuild)
	if err := indexManager.Initialize(); err != nil {
		fmt.Printf("[DAO] Warning: Failed to initialize index: %v\n", err)
	}

	return &ItemDAO{
		filename:     filename,
		indexManager: indexManager,
	}
}

func (dao *ItemDAO) Write(text string) error {
	// Append entry and get result with recordID and offset
	result, err := persistence.AppendItem(dao.filename, text)
	if err != nil {
		return err
	}

	// Update index with the new record
	if err := dao.indexManager.Insert(result.RecordID, result.Offset); err != nil {
		fmt.Printf("[DAO] Warning: Failed to update index: %v\n", err)
	}

	// Save index to disk
	if err := dao.indexManager.Save(); err != nil {
		fmt.Printf("[DAO] Warning: Failed to save index: %v\n", err)
	}

	return nil
}

func (dao *ItemDAO) Read() ([]persistence.Item, error) {
	return persistence.ReadAllItems(dao.filename)
}

func (dao *ItemDAO) Print() (string, error) {
	return persistence.PrintItemBinaryFile(dao.filename)
}

// GetByID retrieves an item by its record ID using the index
func (dao *ItemDAO) GetByID(recordID uint32) (*persistence.Item, error) {
	return dao.indexManager.GetRecordByID(recordID)
}

// RebuildIndex rebuilds the index from scratch by scanning the data file
func (dao *ItemDAO) RebuildIndex() error {
	return dao.indexManager.RebuildIndex()
}

// PrintIndex prints the B+ tree structure for debugging
func (dao *ItemDAO) PrintIndex() {
	dao.indexManager.PrintTree()
}

// Delete marks a record as deleted by setting its tombstone flag
// Returns the name of the deleted item
func (dao *ItemDAO) Delete(recordID uint32) (string, error) {
	// Get the record first to get its name
	item, err := dao.GetByID(recordID)
	if err != nil {
		return "", fmt.Errorf("record ID %d not found: %w", recordID, err)
	}

	// Get the offset from the index
	offset, found := dao.indexManager.GetOffset(recordID)
	if !found {
		return "", fmt.Errorf("record ID %d not found in index", recordID)
	}

	fmt.Printf("[DAO] Marking record %d (%s) as deleted at offset %d\n", recordID, item.Name, offset)

	// Open file for read/write
	file, err := os.OpenFile(dao.filename, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	// Seek to the tombstone byte (first byte of the record)
	if _, err := file.Seek(offset, 0); err != nil {
		return "", fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	// Write tombstone flag (1 = deleted)
	if _, err := file.Write([]byte{1}); err != nil {
		return "", fmt.Errorf("failed to write tombstone flag: %w", err)
	}

	fmt.Printf("[DAO] Successfully marked record %d (%s) as deleted\n", recordID, item.Name)
	return item.Name, nil
}

// DeleteAllFiles deletes the data file and index file
func (dao *ItemDAO) DeleteAllFiles() error {
	// Delete data file
	if err := os.Remove(dao.filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete data file: %w", err)
	}

	// Delete index file
	indexFilename := dao.filename + ".idx"
	if err := os.Remove(indexFilename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete index file: %w", err)
	}

	fmt.Printf("[DAO] Deleted files: %s and %s\n", dao.filename, indexFilename)
	return nil
}
