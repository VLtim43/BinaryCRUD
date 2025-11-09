package dao

import (
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"sync"
)

// ItemDAO manages the items binary file
type ItemDAO struct {
	filePath string
	mu       sync.Mutex // Protects concurrent writes to the binary file
}

// NewItemDAO creates a new ItemDAO instance
func NewItemDAO(filePath string) *ItemDAO {
	return &ItemDAO{
		filePath: filePath,
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
// Item structure: [ID][0x1F][nameSize][name][0x1F][price][0x1E]
// ID is auto-assigned by AppendEntry
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

	// Append the entry (ID auto-assigned and record separator added)
	err = utils.AppendEntry(file, entry)
	if err != nil {
		return fmt.Errorf("failed to append item: %w", err)
	}

	return nil
}

// Read retrieves an item by ID using sequential search
// Returns (id, name, priceInCents, error)
func (dao *ItemDAO) Read(id uint64) (uint64, string, uint64, error) {
	// Ensure file exists
	if err := dao.ensureFileExists(); err != nil {
		return 0, "", 0, err
	}

	// Open file for reading
	file, err := os.Open(dao.filePath)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to open item file: %w", err)
	}
	defer file.Close()

	// Use sequential finder to locate the entry
	entryData, err := utils.FindByIDSequential(file, id)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to find item: %w", err)
	}

	// Parse the entry: [ID(2)][0x1F][nameSize(2)][name][0x1F][price(4)]
	offset := 0

	// Read ID
	entryID, offset, err := utils.ReadFixedNumber(utils.IDSize, entryData, offset)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to read ID: %w", err)
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
