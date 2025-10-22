package main

import (
	"BinaryCRUD/backend/dao"
	"BinaryCRUD/backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// App struct
type App struct {
	ctx     context.Context
	itemDAO *dao.ItemDAO
}

// ItemDTO represents an item with its ID and name for frontend consumption
type ItemDTO struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		itemDAO: dao.NewItemDAO("data/items.bin"),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// AddItem writes an item to the binary file
func (a *App) AddItem(text string) error {
	return a.itemDAO.Write(text)
}

// GetItems reads items from the binary file and returns them with IDs
func (a *App) GetItems() ([]ItemDTO, error) {
	// TODO: Implement when Read method is added to ItemDAO
	return []ItemDTO{}, nil
}

// PrintBinaryFile prints the binary file to the application console
func (a *App) PrintBinaryFile() error {
	output, err := a.itemDAO.Print()
	if err != nil {
		return err
	}

	// Print to application console (same as debug logs)
	fmt.Println(output)

	return nil
}

// GetItemByID retrieves an item by its record ID using the B+ tree index
func (a *App) GetItemByID(recordID uint32) (string, error) {
	// TODO: Implement when GetByID method is added to ItemDAO
	return "", fmt.Errorf("not yet implemented")
}

// DeleteItem marks an item as deleted by setting its tombstone flag
// Returns the name of the deleted item
func (a *App) DeleteItem(recordID uint32) (string, error) {
	// TODO: Implement when Delete method is added to ItemDAO
	return "", fmt.Errorf("not yet implemented")
}

// RebuildIndex rebuilds the B+ tree index from scratch
func (a *App) RebuildIndex() error {
	// TODO: Implement when index functionality is added
	utils.DebugPrint("RebuildIndex: Not yet implemented")
	return nil
}

// PrintIndex prints the B+ tree structure to the console (for debugging)
func (a *App) PrintIndex() {
	// TODO: Implement when index functionality is added
	utils.DebugPrint("PrintIndex: Not yet implemented")
}

// DeleteAllFiles deletes all files in the data folder
func (a *App) DeleteAllFiles() error {
	dataDir := "data"

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		utils.DebugPrint("Data directory does not exist: %s", dataDir)
		return nil
	}

	// Read all entries in the data directory
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	// Delete each file
	deletedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := fmt.Sprintf("%s/%s", dataDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				utils.DebugPrint("Failed to delete %s: %v", filePath, err)
			} else {
				utils.DebugPrint("Deleted", filePath)
				deletedCount++
			}
		}
	}

	utils.DebugPrint("Deleted %d files from %s", deletedCount, dataDir)
	return nil
}

// InventoryData represents the JSON structure for inventory population
type InventoryData struct {
	Items []string `json:"items"`
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

	for _, itemName := range inventory.Items {
		if itemName == "" {
			errorCount++
			continue
		}

		if err := a.itemDAO.Write(itemName); err != nil {
			utils.DebugPrint("Failed to add item '%s': %v", itemName, err)
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
