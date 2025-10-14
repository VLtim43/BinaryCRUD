package main

import (
	"BinaryCRUD/backend/dao"
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx     context.Context
	itemDAO *dao.ItemDAO
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

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// AddItem writes an item to the binary file
func (a *App) AddItem(text string) error {
	return a.itemDAO.Write(text)
}

// GetItems reads items from the binary file
func (a *App) GetItems() ([]string, error) {
	items, err := a.itemDAO.Read()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names, nil
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

// ValidateBinaryFile validates the structure of the binary file
func (a *App) ValidateBinaryFile() (string, error) {
	result, err := a.itemDAO.Validate()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Valid: %v\nHeader Count: %d\nActual Records: %d\nErrors: %v\nWarnings: %v",
		result.Valid, result.HeaderCount, result.ActualRecords, result.Errors, result.Warnings), nil
}
