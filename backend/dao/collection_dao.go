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
	// Check if file already exists
	if _, err := os.Stat(dao.filePath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Create the file
	file, err := utils.CreateFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to create collection file: %w", err)
	}
	defer file.Close()

	// Write empty header (0 entities, 0 tombstones, nextId=0)
	header, err := utils.WriteHeader(0, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// Write creates a new collection entry
// Binary format: [ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)][itemID2(2)]...[0x1E]
func (dao *CollectionDAO) Write(ownerOrName string, totalPrice uint64, itemIDs []uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	// Build entry without ID: [0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)]...

	// Separator before nameSize
	sep1, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Name size (2 bytes)
	nameSize := len(ownerOrName)
	nameSizeBytes, err := utils.WriteFixedNumber(2, uint64(nameSize))
	if err != nil {
		return fmt.Errorf("failed to write name size: %w", err)
	}

	// Name (variable length)
	nameBytes, err := utils.WriteVariable(ownerOrName)
	if err != nil {
		return fmt.Errorf("failed to write name: %w", err)
	}

	// Separator before total price
	sep2, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Total price (4 bytes)
	totalPriceBytes, err := utils.WriteFixedNumber(4, totalPrice)
	if err != nil {
		return fmt.Errorf("failed to write total price: %w", err)
	}

	// Separator before item count
	sep3, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Item count (4 bytes)
	itemCount := uint64(len(itemIDs))
	itemCountBytes, err := utils.WriteFixedNumber(4, itemCount)
	if err != nil {
		return fmt.Errorf("failed to write item count: %w", err)
	}

	// Separator before item IDs
	sep4, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
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
			return fmt.Errorf("failed to write item ID: %w", err)
		}
		entry = append(entry, itemIDBytes...)
	}

	// Read header to get the next ID
	_, _, nextId, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Seek back to end
	_, err = file.Seek(0, 2)
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	// Get actual append position
	appendPos, err := file.Seek(0, 1)
	if err != nil {
		return fmt.Errorf("failed to get append position: %w", err)
	}

	// Append the entry (ID and tombstone auto-assigned, record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return fmt.Errorf("failed to append collection: %w", err)
	}

	// Add to B+ tree index: ID -> file offset
	dao.tree.Insert(uint64(nextId), appendPos)

	// Save index to disk
	err = dao.tree.Save(dao.indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
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

	// Parse entry: [ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)]...
	parseOffset := 0

	// Read ID
	entryID, parseOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID: %w", err)
	}

	// Read tombstone byte
	if parseOffset >= len(entryData) {
		return nil, fmt.Errorf("entry too short for tombstone")
	}
	tombstone := entryData[parseOffset]
	parseOffset += utils.TombstoneSize

	// Check if deleted
	if tombstone != 0x00 {
		return nil, fmt.Errorf("collection with ID %d is deleted", entryID)
	}

	// Skip unit separator (0x1F)
	parseOffset += 1

	// Read name size
	nameSize, parseOffset, err := utils.ReadFixedNumber(2, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name size: %w", err)
	}

	// Read name
	ownerOrName, parseOffset, err := utils.ReadFixedString(int(nameSize), entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}

	// Skip unit separator (0x1F)
	parseOffset += 1

	// Read total price
	totalPrice, parseOffset, err := utils.ReadFixedNumber(4, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read total price: %w", err)
	}

	// Skip unit separator (0x1F)
	parseOffset += 1

	// Read item count
	itemCount, parseOffset, err := utils.ReadFixedNumber(4, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read item count: %w", err)
	}

	// Skip unit separator (0x1F)
	parseOffset += 1

	// Read item IDs (2 bytes each)
	itemIDs := make([]uint64, itemCount)
	for i := uint64(0); i < itemCount; i++ {
		itemID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, parseOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read item ID %d: %w", i, err)
		}
		itemIDs[i] = itemID
		parseOffset = newOffset
	}

	return &Collection{
		ID:          entryID,
		OwnerOrName: ownerOrName,
		TotalPrice:  totalPrice,
		ItemCount:   itemCount,
		ItemIDs:     itemIDs,
	}, nil
}

// Delete marks a collection as deleted by flipping the tombstone bit
func (dao *CollectionDAO) Delete(id uint64) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	// Read header to get current tombstone count
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Calculate header size to know where entries start
	headerSize := (utils.HeaderFieldSize * 3) + 3 // 15 bytes

	// Read the rest of the file to find the entry
	fileData, err := os.ReadFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Split by record separator to get individual entries
	recordSeparatorByte := []byte(utils.RecordSeparator)[0]
	entries := make([][]byte, 0)
	entryPositions := make([]int64, 0)

	offset := headerSize
	entryStart := offset

	for i := headerSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, fileData[entryStart:i])
			entryPositions = append(entryPositions, int64(entryStart))
			entryStart = i + 1
		}
	}

	// Find the entry with matching ID
	for idx, entryData := range entries {
		if len(entryData) < utils.IDSize {
			continue
		}

		// Read the ID
		entryID, _, err := utils.ReadFixedNumber(utils.IDSize, entryData, 0)
		if err != nil {
			continue
		}

		if entryID == id {
			// Check if already deleted
			if len(entryData) < utils.IDSize+utils.TombstoneSize {
				return fmt.Errorf("malformed entry for ID %d", id)
			}

			tombstone := entryData[utils.IDSize]
			if tombstone != 0x00 {
				return fmt.Errorf("collection with ID %d is already deleted", id)
			}

			// Calculate position of tombstone byte in file
			tombstonePos := entryPositions[idx] + int64(utils.IDSize)

			// Seek to tombstone position
			_, err = file.Seek(tombstonePos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to tombstone: %w", err)
			}

			// Write 0x01 to mark as deleted
			_, err = file.Write([]byte{0x01})
			if err != nil {
				return fmt.Errorf("failed to write tombstone: %w", err)
			}

			// Force write tombstone to disk before updating header
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync tombstone to disk: %w", err)
			}

			// Update header to increment tombstone count
			err = utils.UpdateHeader(file, entitiesCount, tombstoneCount+1, nextId)
			if err != nil {
				return fmt.Errorf("failed to update header: %w", err)
			}

			// Force write header to disk
			err = file.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync header to disk: %w", err)
			}

			// Remove from B+ tree index
			dao.tree.Delete(id)

			// Save updated index to disk
			err = dao.tree.Save(dao.indexPath)
			if err != nil {
				return fmt.Errorf("failed to save index: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("entry with ID %d not found", id)
}

// GetAll retrieves all non-deleted collections using sequential scan
func (dao *CollectionDAO) GetAll() ([]*Collection, error) {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return nil, err
	}

	// Read all file data
	fileData, err := os.ReadFile(dao.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate header size
	headerSize := (utils.HeaderFieldSize * 3) + 3 // 15 bytes

	if len(fileData) <= headerSize {
		return []*Collection{}, nil // No entries yet
	}

	// Split by record separator
	recordSeparatorByte := []byte(utils.RecordSeparator)[0]
	entries := make([][]byte, 0)

	entryStart := headerSize
	for i := headerSize; i < len(fileData); i++ {
		if fileData[i] == recordSeparatorByte {
			entries = append(entries, fileData[entryStart:i])
			entryStart = i + 1
		}
	}

	// Parse each entry
	result := make([]*Collection, 0)

	for _, entryData := range entries {
		if len(entryData) < utils.IDSize+utils.TombstoneSize {
			continue
		}

		offset := 0

		// Read ID
		entryID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
		if err != nil {
			continue
		}

		// Read tombstone
		tombstone := entryData[offset]
		offset += utils.TombstoneSize

		// Skip deleted entries
		if tombstone != 0x00 {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read name size
		nameSize, offset, err := utils.ReadFixedNumber(2, entryData, offset)
		if err != nil {
			continue
		}

		// Read name
		ownerOrName, offset, err := utils.ReadFixedString(int(nameSize), entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read total price
		totalPrice, offset, err := utils.ReadFixedNumber(4, entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read item count
		itemCount, offset, err := utils.ReadFixedNumber(4, entryData, offset)
		if err != nil {
			continue
		}

		// Skip unit separator
		offset += 1

		// Read item IDs
		itemIDs := make([]uint64, itemCount)
		for i := uint64(0); i < itemCount; i++ {
			itemID, newOffset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
			if err != nil {
				break
			}
			itemIDs[i] = itemID
			offset = newOffset
		}

		result = append(result, &Collection{
			ID:          entryID,
			OwnerOrName: ownerOrName,
			TotalPrice:  totalPrice,
			ItemCount:   itemCount,
			ItemIDs:     itemIDs,
		})
	}

	return result, nil
}
