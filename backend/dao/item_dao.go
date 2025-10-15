package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/serialization"
	"fmt"
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
	result, err := serialization.AppendEntry(dao.filename, text)
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

func (dao *ItemDAO) Read() ([]serialization.Item, error) {
	return serialization.ReadAllEntries(dao.filename)
}

func (dao *ItemDAO) Print() (string, error) {
	return serialization.PrintBinaryFile(dao.filename)
}

// GetByID retrieves an item by its record ID using the index
func (dao *ItemDAO) GetByID(recordID uint32) (*serialization.Item, error) {
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
