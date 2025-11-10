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
	indexPath := filePath[:len(filePath)-4] + ".idx" // Replace .bin with .idx

	// Try to load existing index
	tree, err := index.Load(indexPath)
	if err != nil {
		// If load fails, create new empty tree
		tree = index.NewBTree(4)
	}

	return &ItemDAO{
		filePath:  filePath,
		indexPath: indexPath,
		tree:      tree,
	}
}

// ensureFileExists creates the file with empty header if it doesn't exist
func (dao *ItemDAO) ensureFileExists() error {
	// Check if file already exists
	if _, err := os.Stat(dao.filePath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Create the file
	file, err := utils.CreateFile(dao.filePath)
	if err != nil {
		return fmt.Errorf("failed to create item file: %w", err)
	}
	defer file.Close()

	// Write empty header (0 entities, 0 tombstones, nextId=1)
	header, err := utils.WriteHeader(0, 0, 1)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// Write adds an item to the binary file
// Item structure: [ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][price(4)][0x1E]
// ID and tombstone are auto-assigned by AppendEntry (tombstone is 0x00 for active records)
func (dao *ItemDAO) Write(name string, priceInCents uint64) error {
	// Lock to prevent concurrent writes
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open item file: %w", err)
	}
	defer file.Close()

	// Calculate name size
	nameSize := len(name)

	// Build entry without ID: [0x1F][nameSize][name][0x1F][price]

	// Separator before nameSize
	sep1, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Name size (2 bytes - supports names up to 65535 chars)
	nameSizeBytes, err := utils.WriteFixedNumber(2, uint64(nameSize))
	if err != nil {
		return fmt.Errorf("failed to write name size: %w", err)
	}

	// Name (variable length)
	nameBytes, err := utils.WriteVariable(name)
	if err != nil {
		return fmt.Errorf("failed to write name: %w", err)
	}

	// Separator before price
	sep2, err := utils.WriteVariable(utils.UnitSeparator)
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Price (4 bytes - supports prices up to 4,294,967,295 cents)
	priceBytes, err := utils.WriteFixedNumber(4, priceInCents)
	if err != nil {
		return fmt.Errorf("failed to write price: %w", err)
	}

	// Combine all fields
	entry := make([]byte, 0)
	entry = append(entry, sep1...)
	entry = append(entry, nameSizeBytes...)
	entry = append(entry, nameBytes...)
	entry = append(entry, sep2...)
	entry = append(entry, priceBytes...)

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

	// Append the entry (ID auto-assigned and record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return fmt.Errorf("failed to append item: %w", err)
	}

	// Add to index: ID -> file offset
	dao.tree.Insert(uint64(nextId), appendPos)

	// Save index to disk
	err = dao.tree.Save(dao.indexPath)
	if err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// Read retrieves an item by ID using sequential search
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) Read(id uint64) (uint64, string, uint64, error) {
	return dao.readWithIndex(id, false)
}

// ReadWithIndex retrieves an item by ID, optionally using the B+ tree index
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) ReadWithIndex(id uint64, useIndex bool) (uint64, string, uint64, error) {
	return dao.readWithIndex(id, useIndex)
}

// readWithIndex is the internal implementation
func (dao *ItemDAO) readWithIndex(id uint64, useIndex bool) (uint64, string, uint64, error) {
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

	if useIndex {
		// Use B+ tree index for fast lookup
		offset, found := dao.tree.Search(id)
		if !found {
			return 0, "", 0, fmt.Errorf("entry with ID %d not found", id)
		}

		// Seek to the entry position
		_, err = file.Seek(offset, 0)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to seek to offset: %w", err)
		}

		// Read until record separator
		buffer := make([]byte, 4096) // Read in chunks
		n, err := file.Read(buffer)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to read entry: %w", err)
		}

		// Find record separator
		recordSep := []byte(utils.RecordSeparator)[0]
		sepPos := -1
		for i := 0; i < n; i++ {
			if buffer[i] == recordSep {
				sepPos = i
				break
			}
		}

		if sepPos == -1 {
			return 0, "", 0, fmt.Errorf("malformed entry: no record separator")
		}

		entryData = buffer[:sepPos]
	} else {
		// Use sequential finder to locate the entry
		entryData, err = utils.FindByIDSequential(file, id)
		if err != nil {
			return 0, "", 0, fmt.Errorf("failed to find item: %w", err)
		}
	}

	// Parse the entry: [ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][price(4)]
	offset := 0

	// Read ID
	entryID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to read ID: %w", err)
	}

	// Read tombstone byte
	tombstone := entryData[offset]
	offset += utils.TombstoneSize

	// Check if item is deleted
	if tombstone != 0x00 {
		return 0, "", 0, fmt.Errorf("deleted file id %d", entryID)
	}

	// Skip unit separator (0x1F)
	offset += 1

	// Read name size
	nameSize, offset, err := utils.ReadFixedNumber(2, entryData, offset)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to read name size: %w", err)
	}

	// Read name
	name, offset, err := utils.ReadFixedString(int(nameSize), entryData, offset)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to read name: %w", err)
	}

	// Skip unit separator (0x1F)
	offset += 1

	// Read price
	price, _, err := utils.ReadFixedNumber(4, entryData, offset)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to read price: %w", err)
	}

	return entryID, name, price, nil
}

// Delete marks an item as deleted by flipping its tombstone bit
// This is a logical deletion - the data remains in the file but is marked as deleted
func (dao *ItemDAO) Delete(id uint64) error {
	// Lock to prevent concurrent modifications
	dao.mu.Lock()
	defer dao.mu.Unlock()

	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return err
	}

	// Open file for read/write
	file, err := os.OpenFile(dao.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open item file: %w", err)
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
				return fmt.Errorf("deleted file id %d", id)
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

			// Remove from index
			err = dao.tree.Delete(id)
			if err != nil {
				// Not fatal - index might not have this entry
				// Continue anyway
			}

			// Save updated index
			err = dao.tree.Save(dao.indexPath)
			if err != nil {
				return fmt.Errorf("failed to save index: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("entry with ID %d not found", id)
}

// GetIndexTree returns the B+ tree for debugging purposes
func (dao *ItemDAO) GetIndexTree() *index.BTree {
	return dao.tree
}
