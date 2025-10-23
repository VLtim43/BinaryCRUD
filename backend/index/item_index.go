package index

import (
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
)

// ItemIndex manages the B+ tree index for items
type ItemIndex struct {
	tree         *BPlusTree
	indexPath    string
	dataFilePath string
}

// NewItemIndex creates a new item index manager
func NewItemIndex(indexPath string, dataFilePath string) *ItemIndex {
	return &ItemIndex{
		tree:         NewBPlusTree(4), // Order 4 B+ tree
		indexPath:    indexPath,
		dataFilePath: dataFilePath,
	}
}

// Load loads the index from disk
func (idx *ItemIndex) Load() error {
	tree, err := LoadFromFile(idx.indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}
	idx.tree = tree
	return nil
}

// Save persists the index to disk
func (idx *ItemIndex) Save() error {
	return idx.tree.SaveToFile(idx.indexPath)
}

// Insert adds a new item ID and its file offset to the index
func (idx *ItemIndex) Insert(itemID uint32, offset int64) error {
	return idx.tree.Insert(itemID, offset)
}

// Search finds the file offset for a given item ID
func (idx *ItemIndex) Search(itemID uint32) (int64, bool) {
	return idx.tree.Search(itemID)
}

// RebuildFromFile rebuilds the entire index by scanning the data file
func (idx *ItemIndex) RebuildFromFile() error {
	utils.DebugPrint("Rebuilding index from %s", idx.dataFilePath)

	// Create new empty tree
	idx.tree = NewBPlusTree(4)

	// Check if data file exists
	if _, err := os.Stat(idx.dataFilePath); os.IsNotExist(err) {
		utils.DebugPrint("Data file does not exist, index is empty")
		return idx.Save()
	}

	// Open data file
	file, err := os.Open(idx.dataFilePath)
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	// Read header to get past it
	header, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	utils.DebugPrint("Rebuilding index for %d entries", header.EntryCount)

	// Current file offset (start after header)
	currentOffset := int64(utils.HeaderSize)

	// Read records sequentially
	recordCount := 0
	for {
		// Remember the start offset of this record
		recordStartOffset := currentOffset

		// Read until we find a record separator or EOF
		recordBytes := []byte{}
		buf := make([]byte, 1)
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					// End of file reached
					if len(recordBytes) == 0 {
						goto done
					}
					return fmt.Errorf("unexpected EOF in record")
				}
				return fmt.Errorf("failed to read byte: %w", err)
			}
			if n == 0 {
				goto done
			}

			currentOffset++

			if buf[0] == utils.RecordSeparator {
				// End of record
				break
			}

			recordBytes = append(recordBytes, buf[0])
		}

		// Parse the record to extract ID
		if len(recordBytes) < 5 { // At minimum: 4 bytes ID + 1 separator
			continue
		}

		// Extract ID (first 4 bytes, little-endian)
		itemID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24

		// Insert into index
		if err := idx.tree.Insert(itemID, recordStartOffset); err != nil {
			utils.DebugPrint("Warning: failed to insert item ID %d into index: %v", itemID, err)
			continue
		}

		recordCount++
	}

done:
	utils.DebugPrint("Rebuilt index with %d entries", recordCount)

	// Save the index
	return idx.Save()
}

// Print returns a string representation of the index
func (idx *ItemIndex) Print() string {
	return idx.tree.Print()
}

// GetAllEntries returns all entries in the index
func (idx *ItemIndex) GetAllEntries() []Entry {
	return idx.tree.GetAllEntries()
}
