package main

import (
	"BinaryCRUD/backend/dao"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// App struct
type App struct {
	ctx      context.Context
	itemDAO  *dao.ItemDAO
	orderDAO *dao.OrderDAO
	logger   *Logger
}

// ItemDTO represents an item with its ID, name, and price for frontend consumption
type ItemDTO struct {
	ID           uint32 `json:"id"`
	Name         string `json:"name"`
	PriceInCents uint64 `json:"priceInCents"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := NewLogger(1000) // Store up to 1000 log entries

	return &App{
		itemDAO:  dao.NewItemDAO("data/items.bin"),
		orderDAO: dao.NewOrderDAO("data/orders.bin"),
		logger:   logger,
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
	return a.itemDAO.Write(text, priceInCents)
}

// AddOrder writes an order to the binary file with an array of item names
func (a *App) AddOrder(itemNames []string) error {
	return a.orderDAO.Write(itemNames)
}

// GetItems reads items from the binary file and returns them with IDs and prices
func (a *App) GetItems() ([]ItemDTO, error) {
	items, err := a.itemDAO.Read()
	if err != nil {
		return []ItemDTO{}, err
	}

	// Convert map to slice of DTOs
	result := make([]ItemDTO, 0, len(items))
	for id, itemData := range items {
		result = append(result, ItemDTO{
			ID:           id,
			Name:         itemData.Name,
			PriceInCents: itemData.PriceInCents,
		})
	}

	return result, nil
}

// GetOrders reads all orders from the binary file
func (a *App) GetOrders() ([]dao.OrderDTO, error) {
	return a.orderDAO.Read()
}

// GetOrderByID reads a single order by its ID
func (a *App) GetOrderByID(orderID uint32) (*dao.OrderDTO, error) {
	return a.orderDAO.ReadByID(orderID)
}


// GetItemByID retrieves an item by its record ID
func (a *App) GetItemByID(recordID uint32) (ItemDTO, error) {
	// Read all items
	items, err := a.itemDAO.Read()
	if err != nil {
		return ItemDTO{}, fmt.Errorf("failed to read items: %w", err)
	}

	// Look up the item by ID
	itemData, exists := items[recordID]
	if !exists {
		return ItemDTO{}, fmt.Errorf("item with ID %d not found", recordID)
	}

	return ItemDTO{
		ID:           recordID,
		Name:         itemData.Name,
		PriceInCents: itemData.PriceInCents,
	}, nil
}

// DeleteItem marks an item as deleted by setting its tombstone flag
// Returns the name of the deleted item
func (a *App) DeleteItem(recordID uint32) (string, error) {
	return a.itemDAO.Delete(recordID)
}

// DeleteOrder marks an order as deleted by setting its tombstone flag
// Returns the number of items in the deleted order
func (a *App) DeleteOrder(orderID uint32) (int, error) {
	return a.orderDAO.Delete(orderID)
}

// RebuildIndex rebuilds the B+ tree index from scratch
func (a *App) RebuildIndex() error {
	return a.itemDAO.RebuildIndex()
}

// PrintIndex prints the B+ tree structure to the console (for debugging)
func (a *App) PrintIndex() {
	indexStr := a.itemDAO.PrintIndex()
	a.logger.LogPrintln(indexStr)
}

// GetItemByIDWithIndex retrieves an item by its ID using the B+ tree index
func (a *App) GetItemByIDWithIndex(recordID uint32) (ItemDTO, error) {
	itemData, err := a.itemDAO.ReadByIDWithIndex(recordID)
	if err != nil {
		return ItemDTO{}, err
	}
	return ItemDTO{
		ID:           recordID,
		Name:         itemData.Name,
		PriceInCents: itemData.PriceInCents,
	}, nil
}

// DeleteAllFiles deletes all files in the data folder
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

	// Delete each file
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := fmt.Sprintf("%s/%s", dataDir, entry.Name())
			os.Remove(filePath)
		}
	}

	return nil
}

// InventoryData represents the JSON structure for inventory population
type InventoryData struct {
	Items []struct {
		Name         string `json:"name"`
		PriceInCents uint64 `json:"priceInCents"`
	} `json:"items"`
}

// PopulateInventory reads a JSON file and adds all items to the binary file
func (a *App) PopulateInventory(filePath string) (string, error) {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var inventory InventoryData
	if err := json.Unmarshal(data, &inventory); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate that we have items
	if len(inventory.Items) == 0 {
		return "No items found in JSON", nil
	}

	// Add each item
	successCount := 0
	errorCount := 0

	for _, item := range inventory.Items {
		if item.Name == "" {
			errorCount++
			continue
		}

		if err := a.itemDAO.Write(item.Name, item.PriceInCents); err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Return summary
	result := fmt.Sprintf("Added %d items", successCount)
	if errorCount > 0 {
		result += fmt.Sprintf(", %d failed", errorCount)
	}

	return result, nil
}

// GetLogs returns all current log entries
func (a *App) GetLogs() []LogEntry {
	return a.logger.GetLogs()
}

// ClearLogs clears all log entries
func (a *App) ClearLogs() {
	a.logger.Clear()
}
