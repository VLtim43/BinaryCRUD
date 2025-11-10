package main

import (
	"BinaryCRUD/backend/dao"
	"BinaryCRUD/backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// App struct
type App struct {
	ctx          context.Context
	itemDAO      *dao.ItemDAO
	orderDAO     *dao.OrderDAO
	promotionDAO *dao.PromotionDAO
	logger       *Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := NewLogger(1000) // Store up to 1000 log entries

	return &App{
		itemDAO:      dao.NewItemDAO("data/items.bin"),
		orderDAO:     dao.NewOrderDAO("data/orders.bin"),
		promotionDAO: dao.NewPromotionDAO("data/promotions.bin"),
		logger:       logger,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Log("Application started")
}

// AddItem writes an item to the binary file with a price in cents
func (a *App) AddItem(text string, priceInCents uint64) error {
	// Convert item name to hexadecimal for debugging (with spaces between bytes)
	bytes := []byte(text)
	hexParts := make([]string, len(bytes))
	for i, b := range bytes {
		hexParts[i] = fmt.Sprintf("%02x", b)
	}
	hexName := ""
	for i, part := range hexParts {
		if i > 0 {
			hexName += " "
		}
		hexName += part
	}

	// Log debugging information
	a.logger.Log(fmt.Sprintf("[debugging] created item %s [%s]", text, hexName))

	return a.itemDAO.Write(text, priceInCents)
}

// GetItem retrieves an item by ID from the binary file
func (a *App) GetItem(id uint64, useIndex bool) (map[string]any, error) {
	itemID, name, priceInCents, err := a.itemDAO.ReadWithIndex(id, useIndex)
	if err != nil {
		return nil, err
	}

	a.logger.Log(fmt.Sprintf("Read item ID %d using index: %v", id, useIndex))

	return map[string]any{
		"id":           itemID,
		"name":         name,
		"priceInCents": priceInCents,
	}, nil
}

// DeleteItem marks an item as deleted by flipping its tombstone bit
func (a *App) DeleteItem(id uint64) error {
	err := a.itemDAO.Delete(id)
	if err != nil {
		return err
	}

	a.logger.Log(fmt.Sprintf("Deleted item with ID: %d", id))
	return nil
}

// DeleteAllFiles deletes all files in the data folder except .json files
func (a *App) DeleteAllFiles() error {
	dataDir := "data"

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return nil
	}

	// Read all entries in the data directory
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	// Delete each file except .json files
	deletedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := entry.Name()
			filePath := fmt.Sprintf("%s/%s", dataDir, fileName)

			// Skip .json files
			if len(fileName) >= 5 && fileName[len(fileName)-5:] == ".json" {
				a.logger.Log(fmt.Sprintf("Skipping JSON file: %s", fileName))
				continue
			}

			err := os.Remove(filePath)
			if err != nil {
				a.logger.Log(fmt.Sprintf("Failed to delete %s: %v", fileName, err))
			} else {
				a.logger.Log(fmt.Sprintf("Deleted file: %s", fileName))
				deletedCount++
			}
		}
	}

	a.logger.Log(fmt.Sprintf("Deleted %d file(s), skipped .json files", deletedCount))

	// Reload ItemDAO to clear the in-memory index
	a.itemDAO = dao.NewItemDAO("data/items.bin")
	a.logger.Log("Cleared in-memory index")

	return nil
}

// GetLogs returns all current log entries
func (a *App) GetLogs() []LogEntry {
	return a.logger.GetLogs()
}

// ClearLogs clears all log entries
func (a *App) ClearLogs() {
	a.logger.Clear()
}

// ItemEntry represents an item in the JSON file
type ItemEntry struct {
	Name         string `json:"name"`
	PriceInCents uint64 `json:"priceInCents"`
}

// GetIndexContents returns the contents of the B+ tree index for debugging
func (a *App) GetIndexContents() (map[string]any, error) {
	// Get all entries from the tree
	tree := a.itemDAO.GetIndexTree()
	allEntries := tree.GetAll()

	// Convert to sorted slice for display
	type IndexEntry struct {
		ID     uint64 `json:"id"`
		Offset int64  `json:"offset"`
	}

	entries := make([]IndexEntry, 0, len(allEntries))
	for id, offset := range allEntries {
		entries = append(entries, IndexEntry{
			ID:     id,
			Offset: offset,
		})
	}

	// Sort by ID
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})

	a.logger.Log(fmt.Sprintf("Index contains %d entries", len(entries)))

	return map[string]any{
		"count":   len(entries),
		"entries": entries,
	}, nil
}

// PopulateInventory reads items from items.json and adds them to the database
// with delays to ensure safe sequential writes
func (a *App) PopulateInventory() error {
	// Read the JSON file
	jsonPath := "data/items.json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read items.json: %w", err)
	}

	// Parse JSON into slice of items
	var items []ItemEntry
	err = json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("failed to parse items.json: %w", err)
	}

	a.logger.Log(fmt.Sprintf("Starting inventory population with %d items", len(items)))

	// Add each item sequentially with a delay to prevent race conditions
	successCount := 0
	failCount := 0

	for i, item := range items {
		// Add item using the Write method (protected by mutex)
		err := a.itemDAO.Write(item.Name, item.PriceInCents)
		if err != nil {
			a.logger.Log(fmt.Sprintf("Failed to add item %d (%s): %v", i+1, item.Name, err))
			failCount++
			continue
		}

		successCount++
		a.logger.Log(fmt.Sprintf("Added item %d/%d: %s ($%.2f)", i+1, len(items), item.Name, float64(item.PriceInCents)/100))

		// Small delay to ensure file system has time to complete the write
		// This prevents potential file corruption from rapid sequential writes
		time.Sleep(10 * time.Millisecond)
	}

	a.logger.Log(fmt.Sprintf("Inventory population complete: %d succeeded, %d failed", successCount, failCount))

	if failCount > 0 {
		return fmt.Errorf("some items failed to add: %d succeeded, %d failed", successCount, failCount)
	}

	return nil
}

// GetAllItems retrieves all non-deleted items from the database
func (a *App) GetAllItems() ([]map[string]any, error) {
	items, err := a.itemDAO.GetAll()
	if err != nil {
		return nil, err
	}

	// Convert to map format for JSON serialization
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"id":           item.ID,
			"name":         item.Name,
			"priceInCents": item.PriceInCents,
		}
	}

	a.logger.Log(fmt.Sprintf("Retrieved %d items", len(items)))
	return result, nil
}

// CreateOrder creates a new order with the given customer name and item IDs
func (a *App) CreateOrder(customerName string, itemIDs []uint64) (uint64, error) {
	// Validate inputs
	if customerName == "" {
		return 0, fmt.Errorf("customer name cannot be empty")
	}

	if len(itemIDs) == 0 {
		return 0, fmt.Errorf("order must contain at least one item")
	}

	// Calculate total price by reading each item
	var totalPrice uint64
	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.ReadWithIndex(itemID, true)
		if err != nil {
			return 0, fmt.Errorf("failed to read item %d: %w", itemID, err)
		}
		totalPrice += priceInCents
	}

	// Read current header to get next ID
	file, err := os.OpenFile("data/orders.bin", os.O_RDONLY, 0644)
	var nextID uint64 = 1
	if err == nil {
		_, _, id, readErr := utils.ReadHeader(file)
		if readErr == nil {
			nextID = uint64(id)
		}
		file.Close()
	}

	// Write order to orders.bin
	err = a.orderDAO.Write(customerName, totalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	a.logger.Log(fmt.Sprintf("Created order #%d for %s with %d items (total: $%.2f)",
		nextID, customerName, len(itemIDs), float64(totalPrice)/100))

	return nextID, nil
}

// GetOrder retrieves an order by ID
func (a *App) GetOrder(id uint64) (map[string]any, error) {
	order, err := a.orderDAO.Read(id)
	if err != nil {
		return nil, err
	}

	a.logger.Log(fmt.Sprintf("Retrieved order #%d for %s", id, order.OwnerOrName))

	return map[string]any{
		"id":         order.ID,
		"customer":   order.OwnerOrName,
		"totalPrice": order.TotalPrice,
		"itemCount":  order.ItemCount,
		"itemIDs":    order.ItemIDs,
	}, nil
}

// DeleteOrder marks an order as deleted
func (a *App) DeleteOrder(id uint64) error {
	err := a.orderDAO.Delete(id)
	if err != nil {
		return err
	}

	a.logger.Log(fmt.Sprintf("Deleted order #%d", id))
	return nil
}

// CreatePromotion creates a new promotion with the given name and item IDs
func (a *App) CreatePromotion(promotionName string, itemIDs []uint64) (uint64, error) {
	// Validate inputs
	if promotionName == "" {
		return 0, fmt.Errorf("promotion name cannot be empty")
	}

	if len(itemIDs) == 0 {
		return 0, fmt.Errorf("promotion must contain at least one item")
	}

	// Calculate total price by reading each item
	var totalPrice uint64
	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.ReadWithIndex(itemID, true)
		if err != nil {
			return 0, fmt.Errorf("failed to read item %d: %w", itemID, err)
		}
		totalPrice += priceInCents
	}

	// Read current header to get next ID
	file, err := os.OpenFile("data/promotions.bin", os.O_RDONLY, 0644)
	var nextID uint64 = 1
	if err == nil {
		_, _, id, readErr := utils.ReadHeader(file)
		if readErr == nil {
			nextID = uint64(id)
		}
		file.Close()
	}

	// Write promotion to promotions.bin
	err = a.promotionDAO.Write(promotionName, totalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create promotion: %w", err)
	}

	a.logger.Log(fmt.Sprintf("Created promotion #%d: %s with %d items (total: $%.2f)",
		nextID, promotionName, len(itemIDs), float64(totalPrice)/100))

	return nextID, nil
}

// GetPromotion retrieves a promotion by ID
func (a *App) GetPromotion(id uint64) (map[string]any, error) {
	promotion, err := a.promotionDAO.Read(id)
	if err != nil {
		return nil, err
	}

	a.logger.Log(fmt.Sprintf("Retrieved promotion #%d: %s", id, promotion.OwnerOrName))

	return map[string]any{
		"id":         promotion.ID,
		"name":       promotion.OwnerOrName,
		"totalPrice": promotion.TotalPrice,
		"itemCount":  promotion.ItemCount,
		"itemIDs":    promotion.ItemIDs,
	}, nil
}

// DeletePromotion marks a promotion as deleted
func (a *App) DeletePromotion(id uint64) error {
	err := a.promotionDAO.Delete(id)
	if err != nil {
		return err
	}

	a.logger.Log(fmt.Sprintf("Deleted promotion #%d", id))
	return nil
}
