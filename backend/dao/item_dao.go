package dao

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
	"encoding/binary"
	"fmt"
	"os"
)

type ItemDAO struct {
	filePath  string
	index     *index.ItemIndex
	indexPath string
}

// NewItemDAO creates a new ItemDAO instance
func NewItemDAO(filePath string) *ItemDAO {
	indexPath := filePath + ".idx"
	return &ItemDAO{
		filePath:  filePath,
		indexPath: indexPath,
		index:     index.NewItemIndex(indexPath, filePath),
	}
}

// Write adds a new item to the binary file with auto-increment ID
// Creates and initializes the file if it doesn't exist
// Also updates the B+ tree index with the new item
// Item record format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][Price(8)][UnitSeparator][StringLength(2)][UnitSeparator][StringContent][UnitSeparator][RecordSeparator]
func (dao *ItemDAO) Write(itemName string, priceInCents uint64) error {
	// Initialize file if it doesn't exist
	if err := utils.InitializeBinaryFile(dao.filePath); err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}

	// Load index
	if err := dao.index.Load(); err != nil {
		// Silently fail to load index
	}

	// Open file in read-write mode for the entire operation
	file, err := utils.OpenFileForWrite(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read current header
	header, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Get the ID to use for this record
	itemID := header.NextID

	// Get current file size to know where record will be written
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	recordOffset := fileInfo.Size()

	// Build the ID field: [ID(4)][UnitSeparator]
	idBytes, err := utils.BuildFixed(4, uint64(itemID))
	if err != nil {
		return fmt.Errorf("failed to build ID field: %w", err)
	}

	// Build the tombstone field: [Tombstone(1)][UnitSeparator] - 0x00 = active
	tombstoneBytes, err := utils.BuildFixed(1, 0x00)
	if err != nil {
		return fmt.Errorf("failed to build tombstone field: %w", err)
	}

	// Build the price field: [Price(8)][UnitSeparator]
	priceBytes, err := utils.BuildFixed(8, priceInCents)
	if err != nil {
		return fmt.Errorf("failed to build price field: %w", err)
	}

	// Build the item name field: [StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
	nameBytes, err := utils.BuildVariable(itemName)
	if err != nil {
		return fmt.Errorf("failed to build item name: %w", err)
	}

	// Combine ID, tombstone, price, and name bytes
	recordBytes := append(idBytes, tombstoneBytes...)
	recordBytes = append(recordBytes, priceBytes...)
	recordBytes = append(recordBytes, nameBytes...)

	// Seek to end of file to append the new record
	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("failed to seek to end of file: %w", err)
	}

	// Write the record bytes
	if _, err := file.Write(recordBytes); err != nil {
		return fmt.Errorf("failed to write record data: %w", err)
	}

	// Write record separator to mark end of record
	if _, err := file.Write([]byte{utils.RecordSeparator}); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}

	// Update header with new counts
	header.EntryCount++
	header.NextID++

	// Write updated header using the same file handle
	if err := utils.WriteHeader(file, header); err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	// Update index
	if err := dao.index.Insert(itemID, recordOffset); err != nil {
		// Silently fail to insert into index
	}

	// Save index
	if err := dao.index.Save(); err != nil {
		// Silently fail to save index
	}

	return nil
}

// ItemData represents an item with its name and price
type ItemData struct {
	Name         string
	PriceInCents uint64
}

// Read reads all items from the binary file and returns them with their IDs
// Skips records marked as deleted (tombstone = 0x01)
func (dao *ItemDAO) Read() (map[uint32]ItemData, error) {
	items := make(map[uint32]ItemData)

	// Use generic sequential read to get raw record bytes
	records, err := utils.SequentialRead(dao.filePath)
	if err != nil {
		return items, err
	}

	// Parse each record's bytes to extract the ID, price, and item name
	// Format: [ID(4)][UnitSeparator][Tombstone(1)][UnitSeparator][Price(8)][UnitSeparator][StringLength(2)][UnitSeparator][StringContent][UnitSeparator]
	for _, recordBytes := range records {
		if len(recordBytes) < 20 { // Minimum: 4 bytes ID + 1 sep + 1 tombstone + 1 sep + 8 bytes price + 1 sep + 2 bytes length + 1 sep + 0 content + 1 sep
			return items, fmt.Errorf("invalid record: too short")
		}

		// Extract ID (first 4 bytes, little-endian)
		itemID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24

		// Verify unit separator after ID at position 4
		if recordBytes[4] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after ID", itemID)
		}

		// Extract tombstone flag (byte 5)
		tombstone := recordBytes[5]

		// Verify unit separator after tombstone at position 6
		if recordBytes[6] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after tombstone", itemID)
		}

		// Skip deleted records
		if tombstone == 0x01 {
			continue
		}

		price := binary.LittleEndian.Uint64(recordBytes[7:15])

		// Verify unit separator after price at position 15
		if recordBytes[15] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after price", itemID)
		}

		// Extract length (bytes 16-17, little-endian)
		length := uint16(recordBytes[16]) | uint16(recordBytes[17])<<8

		// Verify unit separator after length at position 18
		if recordBytes[18] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing unit separator after length", itemID)
		}

		// Extract content (after fourth separator)
		contentStart := 19
		contentEnd := contentStart + int(length)

		if contentEnd > len(recordBytes) {
			return items, fmt.Errorf("invalid record ID %d: content length mismatch", itemID)
		}

		// Verify final unit separator
		if contentEnd < len(recordBytes) && recordBytes[contentEnd] != utils.UnitSeparator {
			return items, fmt.Errorf("invalid record ID %d: missing final unit separator", itemID)
		}

		itemName := string(recordBytes[contentStart:contentEnd])
		items[itemID] = ItemData{
			Name:         itemName,
			PriceInCents: price,
		}
	}

	return items, nil
}

// ReadByIDWithIndex reads a single item by ID using the B+ tree index
func (dao *ItemDAO) ReadByIDWithIndex(itemID uint32) (ItemData, error) {
	// Load index
	if err := dao.index.Load(); err != nil {
		return ItemData{}, fmt.Errorf("failed to load index: %w", err)
	}

	// Search for offset in index
	offset, found := dao.index.Search(itemID)
	if !found {
		return ItemData{}, fmt.Errorf("item with ID %d not found in index", itemID)
	}

	// Open file
	file, err := os.Open(dao.filePath)
	if err != nil {
		return ItemData{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to the record offset
	if _, err := file.Seek(offset, 0); err != nil {
		return ItemData{}, fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	// Read record until record separator
	recordBytes := []byte{}
	buf := make([]byte, 1)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return ItemData{}, fmt.Errorf("failed to read record: %w", err)
		}
		if n == 0 {
			return ItemData{}, fmt.Errorf("unexpected EOF while reading record")
		}

		if buf[0] == utils.RecordSeparator {
			break
		}

		recordBytes = append(recordBytes, buf[0])
	}

	// Parse the record
	if len(recordBytes) < 20 {
		return ItemData{}, fmt.Errorf("invalid record: too short")
	}

	// Verify ID matches
	recordID := uint32(recordBytes[0]) | uint32(recordBytes[1])<<8 | uint32(recordBytes[2])<<16 | uint32(recordBytes[3])<<24
	if recordID != itemID {
		return ItemData{}, fmt.Errorf("ID mismatch: expected %d, got %d", itemID, recordID)
	}

	// Verify unit separator after ID
	if recordBytes[4] != utils.UnitSeparator {
		return ItemData{}, fmt.Errorf("invalid record: missing separator after ID")
	}

	// Extract tombstone flag
	tombstone := recordBytes[5]

	// Verify unit separator after tombstone
	if recordBytes[6] != utils.UnitSeparator {
		return ItemData{}, fmt.Errorf("invalid record: missing separator after tombstone")
	}

	// Check if record is deleted
	if tombstone == 0x01 {
		return ItemData{}, fmt.Errorf("item with ID %d has been deleted", itemID)
	}

	// Extract price (bytes 7-14, little-endian uint64)
	price := binary.LittleEndian.Uint64(recordBytes[7:15])

	// Verify unit separator after price
	if recordBytes[15] != utils.UnitSeparator {
		return ItemData{}, fmt.Errorf("invalid record: missing separator after price")
	}

	// Extract length
	length := uint16(recordBytes[16]) | uint16(recordBytes[17])<<8

	// Verify unit separator after length
	if recordBytes[18] != utils.UnitSeparator {
		return ItemData{}, fmt.Errorf("invalid record: missing separator after length")
	}

	// Extract content
	contentStart := 19
	contentEnd := contentStart + int(length)

	if contentEnd > len(recordBytes) {
		return ItemData{}, fmt.Errorf("invalid record: content length mismatch")
	}

	itemName := string(recordBytes[contentStart:contentEnd])
	return ItemData{
		Name:         itemName,
		PriceInCents: price,
	}, nil
}

// Delete marks an item as deleted by setting its tombstone flag to 0x01
// Uses the B+ tree index to locate the record efficiently
func (dao *ItemDAO) Delete(itemID uint32) (string, error) {
	// Load index
	if err := dao.index.Load(); err != nil {
		return "", fmt.Errorf("failed to load index: %w", err)
	}

	// Search for offset in index
	offset, found := dao.index.Search(itemID)
	if !found {
		return "", fmt.Errorf("item with ID %d not found in index", itemID)
	}

	// Open file for reading and writing
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to the record offset
	if _, err := file.Seek(offset, 0); err != nil {
		return "", fmt.Errorf("failed to seek to offset %d: %w", offset, err)
	}

	// Read record until record separator to get full record
	recordBytes := []byte{}
	buf := make([]byte, 1)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return "", fmt.Errorf("failed to read record: %w", err)
		}
		if n == 0 {
			return "", fmt.Errorf("unexpected EOF while reading record")
		}

		if buf[0] == utils.RecordSeparator {
			break
		}

		recordBytes = append(recordBytes, buf[0])
	}

	// Parse the record to validate and extract item name
	if len(recordBytes) < 20 {
		return "", fmt.Errorf("invalid record: too short")
	}

	// Verify ID matches
	recordID := binary.LittleEndian.Uint32(recordBytes[0:4])
	if recordID != itemID {
		return "", fmt.Errorf("ID mismatch: expected %d, got %d", itemID, recordID)
	}

	// Verify unit separator after ID
	if recordBytes[4] != utils.UnitSeparator {
		return "", fmt.Errorf("invalid record: missing separator after ID")
	}

	// Check current tombstone status
	tombstone := recordBytes[5]
	if tombstone == 0x01 {
		return "", fmt.Errorf("item with ID %d is already deleted", itemID)
	}

	// Verify unit separator after tombstone
	if recordBytes[6] != utils.UnitSeparator {
		return "", fmt.Errorf("invalid record: missing separator after tombstone")
	}

	// Verify unit separator after price (at position 15)
	if recordBytes[15] != utils.UnitSeparator {
		return "", fmt.Errorf("invalid record: missing separator after price")
	}

	// Extract item name before deletion (after price field)
	length := binary.LittleEndian.Uint16(recordBytes[16:18])
	contentStart := 19
	contentEnd := contentStart + int(length)
	if contentEnd > len(recordBytes) {
		return "", fmt.Errorf("invalid record: content length mismatch")
	}
	itemName := string(recordBytes[contentStart:contentEnd])

	// Seek to the tombstone byte position (offset + 4 bytes ID + 1 byte separator = offset + 5)
	tombstoneOffset := offset + 5
	if _, err := file.Seek(tombstoneOffset, 0); err != nil {
		return "", fmt.Errorf("failed to seek to tombstone position: %w", err)
	}

	// Write 0x01 to mark as deleted
	if _, err := file.Write([]byte{0x01}); err != nil {
		return "", fmt.Errorf("failed to write tombstone flag: %w", err)
	}

	// Increment tombstone count in header
	if err := utils.IncrementTombstoneCount(dao.filePath); err != nil {
		// Silently fail to increment tombstone count
	}

	return itemName, nil
}

// RebuildIndex rebuilds the B+ tree index from the data file
func (dao *ItemDAO) RebuildIndex() error {
	return dao.index.RebuildFromFile()
}

// PrintIndex returns a string representation of the B+ tree index structure
func (dao *ItemDAO) PrintIndex() string {
	if err := dao.index.Load(); err != nil {
		return fmt.Sprintf("Failed to load index: %v", err)
	}
	return dao.index.Print()
}

