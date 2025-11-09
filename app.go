package main

import (
	"BinaryCRUD/backend/dao"
	"context"
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

// GetLogs returns all current log entries
func (a *App) GetLogs() []LogEntry {
	return a.logger.GetLogs()
}

// ClearLogs clears all log entries
func (a *App) ClearLogs() {
	a.logger.Clear()
}
