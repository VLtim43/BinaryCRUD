package main

import (
	"BinaryCRUD/backend/dao"
	"BinaryCRUD/backend/serialization"
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
}

// ItemDTO represents an item with its ID and name for frontend consumption
type ItemDTO struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		itemDAO:  dao.NewItemDAO("data/items.bin"),
		orderDAO: dao.NewOrderDAO("data/orders.bin"),
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
	items, err := a.itemDAO.Read()
	if err != nil {
		return nil, err
	}

	dtos := make([]ItemDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, ItemDTO{
			ID:   item.RecordID,
			Name: item.Name,
		})
	}
	return dtos, nil
}

// PrintBinaryFile prints the binary file to the application console
func (a *App) PrintBinaryFile() error {
	output, err := a.itemDAO.Print()
	if err != nil {
		return err
	}

	// Print to application console (same as debug logs)
	fmt.Println("\n" + output)

	return nil
}

// GetItemByID retrieves an item by its record ID using the B+ tree index
func (a *App) GetItemByID(recordID uint32) (string, error) {
	item, err := a.itemDAO.GetByID(recordID)
	if err != nil {
		return "", err
	}

	// Include deletion status in the response
	if item.Tombstone {
		return fmt.Sprintf("%s (deleted)", item.Name), nil
	}

	return item.Name, nil
}

// DeleteItem marks an item as deleted by setting its tombstone flag
// Returns the name of the deleted item
func (a *App) DeleteItem(recordID uint32) (string, error) {
	return a.itemDAO.Delete(recordID)
}

// RebuildIndex rebuilds the B+ tree index from scratch
func (a *App) RebuildIndex() error {
	return a.itemDAO.RebuildIndex()
}

// PrintIndex prints the B+ tree structure to the console (for debugging)
func (a *App) PrintIndex() {
	a.itemDAO.PrintIndex()
}

// DeleteAllFiles deletes all files in the data folder
func (a *App) DeleteAllFiles() error {
	dataDir := "data"

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("[DeleteAllFiles] Data directory does not exist: %s\n", dataDir)
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
				fmt.Printf("[DeleteAllFiles] Failed to delete %s: %v\n", filePath, err)
			} else {
				fmt.Printf("[DeleteAllFiles] Deleted: %s\n", filePath)
				deletedCount++
			}
		}
	}

	fmt.Printf("[DeleteAllFiles] Deleted %d files from %s\n", deletedCount, dataDir)
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
			fmt.Printf("Failed to add item '%s': %v\n", itemName, err)
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

// OrderItemDTO represents an item in an order for frontend consumption
type OrderItemDTO struct {
	ItemID   uint32 `json:"itemId"`
	Quantity uint32 `json:"quantity"`
}

// CreateOrder creates a new order with the given items
func (a *App) CreateOrder(items []OrderItemDTO) error {
	// Validate that we have items
	if len(items) == 0 {
		return fmt.Errorf("cannot create empty order")
	}

	// Convert DTOs to domain objects
	orderItems := make([]serialization.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = serialization.OrderItem{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		}
	}

	return a.orderDAO.Write(orderItems)
}

// PrintOrderBinaryFile prints the order binary file to the application console
func (a *App) PrintOrderBinaryFile() error {
	output, err := a.orderDAO.Print()
	if err != nil {
		return err
	}

	// Print to application console (same as debug logs)
	fmt.Println("\n" + output)

	return nil
}

