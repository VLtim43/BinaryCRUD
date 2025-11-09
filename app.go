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
	ctx     context.Context
	itemDAO *dao.ItemDAO
	logger  *Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := NewLogger(1000) // Store up to 1000 log entries

	return &App{
		itemDAO: dao.NewItemDAO("data/items.bin"),
		logger:  logger,
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
