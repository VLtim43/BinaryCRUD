package main

import (
	"BinaryCRUD/backend/dao"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// App struct
type App struct {
	ctx               context.Context
	itemDAO           *dao.ItemDAO
	orderDAO          *dao.OrderDAO
	promotionDAO      *dao.PromotionDAO
	orderPromotionDAO *dao.OrderPromotionDAO
	logger            *Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := NewLogger(1000) // Store up to 1000 log entries

	return &App{
		itemDAO:           dao.NewItemDAO("data/items.bin"),
		orderDAO:          dao.NewOrderDAO("data/orders.bin"),
		promotionDAO:      dao.NewPromotionDAO("data/promotions.bin"),
		orderPromotionDAO: dao.NewOrderPromotionDAO("data/order_promotions.bin"),
		logger:            logger,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("Application started")
}

// calculateTotalPrice calculates the total price of items by reading each item's price
func (a *App) calculateTotalPrice(itemIDs []uint64) (uint64, error) {
	var totalPrice uint64
	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.ReadWithIndex(itemID, true)
		if err != nil {
			return 0, fmt.Errorf("failed to read item %d: %w", itemID, err)
		}
		totalPrice += priceInCents
	}
	return totalPrice, nil
}

// AddItem writes an item to the binary file with a price in cents and returns the assigned ID
func (a *App) AddItem(text string, priceInCents uint64) (uint64, error) {
	// Convert item name to hexadecimal for debugging (with spaces between bytes)
	var hexName strings.Builder
	for i, b := range []byte(text) {
		if i > 0 {
			hexName.WriteString(" ")
		}
		hexName.WriteString(fmt.Sprintf("%02x", b))
	}

	// Write item and get assigned ID
	assignedID, err := a.itemDAO.Write(text, priceInCents)
	if err != nil {
		return 0, err
	}

	// Log debugging information with assigned ID
	a.logger.Debug(fmt.Sprintf("Created item #%d: %s [hex: %s]", assignedID, text, hexName.String()))

	return assignedID, nil
}

// GetItem retrieves an item by ID from the binary file (uses index with automatic fallback)
func (a *App) GetItem(id uint64, useIndex bool) (map[string]any, error) {
	itemID, name, priceInCents, err := a.itemDAO.ReadWithIndex(id, useIndex)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Read item ID %d", id))

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

	a.logger.Info(fmt.Sprintf("Deleted item with ID: %d", id))
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
			if strings.HasSuffix(fileName, ".json") {
				a.logger.Debug(fmt.Sprintf("Skipping JSON file: %s", fileName))
				continue
			}

			err := os.Remove(filePath)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Failed to delete %s: %v", fileName, err))
			} else {
				a.logger.Info(fmt.Sprintf("Deleted file: %s", fileName))
				deletedCount++
			}
		}
	}

	a.logger.Info(fmt.Sprintf("Deleted %d file(s), skipped .json files", deletedCount))

	// Reload ItemDAO to clear the in-memory index
	a.itemDAO = dao.NewItemDAO("data/items.bin")
	a.logger.Info("Cleared in-memory index")

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

// PromotionEntry represents a promotion in the JSON file
type PromotionEntry struct {
	Name    string   `json:"name"`
	ItemIDs []uint64 `json:"itemIDs"`
}

// OrderEntry represents an order in the JSON file
type OrderEntry struct {
	Owner   string   `json:"owner"`
	ItemIDs []uint64 `json:"itemIDs"`
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

	a.logger.Info(fmt.Sprintf("Index contains %d entries", len(entries)))

	return map[string]any{
		"count":   len(entries),
		"entries": entries,
	}, nil
}

// PopulateInventory reads items and promotions from JSON files and adds them to the database
// with delays to ensure safe sequential writes
func (a *App) PopulateInventory() error {
	jsonPath := "data/seed/items.json"
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

	a.logger.Info(fmt.Sprintf("Starting data population with %d items", len(items)))

	// Add each item sequentially with a delay to prevent race conditions
	itemSuccessCount := 0
	itemFailCount := 0

	for i, item := range items {
		// Add item using the Write method (protected by mutex)
		_, err := a.itemDAO.Write(item.Name, item.PriceInCents)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add item %d (%s): %v", i+1, item.Name, err))
			itemFailCount++
			continue
		}

		itemSuccessCount++
		a.logger.Info(fmt.Sprintf("Added item %d/%d: %s ($%.2f)", i+1, len(items), item.Name, float64(item.PriceInCents)/100))
	}

	a.logger.Info(fmt.Sprintf("Items population complete: %d succeeded, %d failed", itemSuccessCount, itemFailCount))

	promoJsonPath := "data/seed/promotions.json"
	promoData, err := os.ReadFile(promoJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read promotions.json: %w", err)
	}

	// Parse JSON into slice of promotions
	var promotions []PromotionEntry
	err = json.Unmarshal(promoData, &promotions)
	if err != nil {
		return fmt.Errorf("failed to parse promotions.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting promotion population with %d promotions", len(promotions)))

	// Add each promotion sequentially
	promoSuccessCount := 0
	promoFailCount := 0

	for i, promo := range promotions {
		// Calculate total price for promotion
		totalPrice := uint64(0)
		for _, itemID := range promo.ItemIDs {
			// Get item to calculate price
			_, _, priceInCents, err := a.itemDAO.Read(itemID)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Item ID %d in promotion '%s' not found, skipping price calculation", itemID, promo.Name))
				continue
			}
			totalPrice += priceInCents
		}

		// Create promotion using CollectionDAO
		_, err := a.promotionDAO.Write(promo.Name, totalPrice, promo.ItemIDs)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add promotion %d (%s): %v", i+1, promo.Name, err))
			promoFailCount++
			continue
		}

		promoSuccessCount++
		a.logger.Info(fmt.Sprintf("Added promotion %d/%d: %s with %d items ($%.2f)",
			i+1, len(promotions), promo.Name, len(promo.ItemIDs), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Promotions population complete: %d succeeded, %d failed", promoSuccessCount, promoFailCount))

	orderJsonPath := "data/seed/orders.json"
	orderData, err := os.ReadFile(orderJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read orders.json: %w", err)
	}

	// Parse JSON into slice of orders
	var orders []OrderEntry
	err = json.Unmarshal(orderData, &orders)
	if err != nil {
		return fmt.Errorf("failed to parse orders.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting orders population with %d orders", len(orders)))

	// Add each order sequentially
	orderSuccessCount := 0
	orderFailCount := 0

	for i, order := range orders {
		totalPrice := uint64(0)
		validItems := []uint64{}

		// Calculate total price
		for _, itemID := range order.ItemIDs {
			_, _, priceInCents, err := a.itemDAO.Read(itemID)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Item ID %d in order '%s' not found, skipping price calculation", itemID, order.Owner))
				continue
			}
			totalPrice += priceInCents
			validItems = append(validItems, itemID)
		}

		if len(validItems) == 0 {
			a.logger.Warn(fmt.Sprintf("Order %d (%s) has no valid items, skipping", i+1, order.Owner))
			orderFailCount++
			continue
		}

		_, err := a.orderDAO.Write(order.Owner, totalPrice, validItems)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add order %d (%s): %v", i+1, order.Owner, err))
			orderFailCount++
			continue
		}

		orderSuccessCount++
		a.logger.Info(fmt.Sprintf("Added order %d/%d: %s with %d items ($%.2f)",
			i+1, len(orders), order.Owner, len(validItems), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Orders population complete: %d succeeded, %d failed", orderSuccessCount, orderFailCount))

	// Final summary
	totalSuccess := itemSuccessCount + promoSuccessCount + orderSuccessCount
	totalFail := itemFailCount + promoFailCount + orderFailCount

	a.logger.Info(fmt.Sprintf("Total population complete: %d items + %d promotions + %d orders = %d total (%d failed)",
		itemSuccessCount, promoSuccessCount, orderSuccessCount, totalSuccess, totalFail))

	if totalFail > 0 {
		return fmt.Errorf("some entries failed to add: %d succeeded, %d failed", totalSuccess, totalFail)
	}

	return nil
}

// PopulateItems reads items from JSON and adds them to the database
func (a *App) PopulateItems() error {
	jsonPath := "data/seed/items.json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read items.json: %w", err)
	}

	var items []ItemEntry
	err = json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("failed to parse items.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting items population with %d items", len(items)))

	successCount := 0
	failCount := 0

	for i, item := range items {
		_, err := a.itemDAO.Write(item.Name, item.PriceInCents)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add item %d (%s): %v", i+1, item.Name, err))
			failCount++
			continue
		}

		successCount++
		a.logger.Info(fmt.Sprintf("Added item %d/%d: %s ($%.2f)", i+1, len(items), item.Name, float64(item.PriceInCents)/100))
	}

	a.logger.Info(fmt.Sprintf("Items population complete: %d succeeded, %d failed", successCount, failCount))

	if failCount > 0 {
		return fmt.Errorf("%d items failed to add", failCount)
	}

	return nil
}

// PopulatePromotions reads promotions from JSON and adds them to the database
func (a *App) PopulatePromotions() error {
	jsonPath := "data/seed/promotions.json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read promotions.json: %w", err)
	}

	var promotions []PromotionEntry
	err = json.Unmarshal(data, &promotions)
	if err != nil {
		return fmt.Errorf("failed to parse promotions.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting promotions population with %d promotions", len(promotions)))

	successCount := 0
	failCount := 0

	for i, promo := range promotions {
		totalPrice := uint64(0)
		for _, itemID := range promo.ItemIDs {
			_, _, priceInCents, err := a.itemDAO.Read(itemID)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Item ID %d in promotion '%s' not found, skipping price calculation", itemID, promo.Name))
				continue
			}
			totalPrice += priceInCents
		}

		_, err := a.promotionDAO.Write(promo.Name, totalPrice, promo.ItemIDs)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add promotion %d (%s): %v", i+1, promo.Name, err))
			failCount++
			continue
		}

		successCount++
		a.logger.Info(fmt.Sprintf("Added promotion %d/%d: %s with %d items ($%.2f)",
			i+1, len(promotions), promo.Name, len(promo.ItemIDs), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Promotions population complete: %d succeeded, %d failed", successCount, failCount))

	if failCount > 0 {
		return fmt.Errorf("%d promotions failed to add", failCount)
	}

	return nil
}

// PopulateOrders reads orders from JSON and adds them to the database
func (a *App) PopulateOrders() error {
	jsonPath := "data/seed/orders.json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read orders.json: %w", err)
	}

	var orders []OrderEntry
	err = json.Unmarshal(data, &orders)
	if err != nil {
		return fmt.Errorf("failed to parse orders.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting orders population with %d orders", len(orders)))

	successCount := 0
	failCount := 0

	for i, order := range orders {
		totalPrice := uint64(0)
		validItems := []uint64{}

		// Calculate total price
		for _, itemID := range order.ItemIDs {
			_, _, priceInCents, err := a.itemDAO.Read(itemID)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Item ID %d in order '%s' not found, skipping price calculation", itemID, order.Owner))
				continue
			}
			totalPrice += priceInCents
			validItems = append(validItems, itemID)
		}

		if len(validItems) == 0 {
			a.logger.Warn(fmt.Sprintf("Order %d (%s) has no valid items, skipping", i+1, order.Owner))
			failCount++
			continue
		}

		_, err := a.orderDAO.Write(order.Owner, totalPrice, validItems)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add order %d (%s): %v", i+1, order.Owner, err))
			failCount++
			continue
		}

		successCount++
		a.logger.Info(fmt.Sprintf("Added order %d/%d: %s with %d items ($%.2f)",
			i+1, len(orders), order.Owner, len(validItems), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Orders population complete: %d succeeded, %d failed", successCount, failCount))

	if failCount > 0 {
		return fmt.Errorf("%d orders failed to add", failCount)
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

	a.logger.Info(fmt.Sprintf("Retrieved %d items", len(items)))
	return result, nil
}

// GetAllOrders retrieves all orders
func (a *App) GetAllOrders() ([]map[string]any, error) {
	orders, err := a.orderDAO.GetAll()
	if err != nil {
		return nil, err
	}

	// Convert to map format for JSON serialization
	result := make([]map[string]any, len(orders))
	for i, order := range orders {
		result[i] = map[string]any{
			"id":          order.ID,
			"customerName": order.OwnerOrName,
			"totalPrice":  order.TotalPrice,
			"itemCount":   order.ItemCount,
			"itemIDs":     order.ItemIDs,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d orders", len(orders)))
	return result, nil
}

// GetAllPromotions retrieves all promotions
func (a *App) GetAllPromotions() ([]map[string]any, error) {
	promotions, err := a.promotionDAO.GetAll()
	if err != nil {
		return nil, err
	}

	// Convert to map format for JSON serialization
	result := make([]map[string]any, len(promotions))
	for i, promotion := range promotions {
		result[i] = map[string]any{
			"id":          promotion.ID,
			"name":        promotion.OwnerOrName,
			"totalPrice":  promotion.TotalPrice,
			"itemCount":   promotion.ItemCount,
			"itemIDs":     promotion.ItemIDs,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d promotions", len(promotions)))
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
	totalPrice, err := a.calculateTotalPrice(itemIDs)
	if err != nil {
		return 0, err
	}

	// Write order to orders.bin and get assigned ID
	assignedID, err := a.orderDAO.Write(customerName, totalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Created order #%d for %s with %d items (total: $%.2f)",
		assignedID, customerName, len(itemIDs), float64(totalPrice)/100))

	return assignedID, nil
}

// GetOrder retrieves an order by ID
func (a *App) GetOrder(id uint64) (map[string]any, error) {
	order, err := a.orderDAO.Read(id)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Retrieved order #%d for %s", id, order.OwnerOrName))

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

	a.logger.Info(fmt.Sprintf("Deleted order #%d", id))
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
	totalPrice, err := a.calculateTotalPrice(itemIDs)
	if err != nil {
		return 0, err
	}

	// Write promotion to promotions.bin and get assigned ID
	assignedID, err := a.promotionDAO.Write(promotionName, totalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create promotion: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Created promotion #%d: %s with %d items (total: $%.2f)",
		assignedID, promotionName, len(itemIDs), float64(totalPrice)/100))

	return assignedID, nil
}

// GetPromotion retrieves a promotion by ID
func (a *App) GetPromotion(id uint64) (map[string]any, error) {
	promotion, err := a.promotionDAO.Read(id)
	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprintf("Retrieved promotion #%d: %s", id, promotion.OwnerOrName))

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

	a.logger.Info(fmt.Sprintf("Deleted promotion #%d", id))
	return nil
}

// ApplyPromotionToOrder applies a promotion to an order (N:N relationship)
func (a *App) ApplyPromotionToOrder(orderID, promotionID uint64) error {
	// Validate order exists
	_, err := a.orderDAO.Read(orderID)
	if err != nil {
		return fmt.Errorf("failed to read order: %w", err)
	}

	// Validate promotion exists
	_, err = a.promotionDAO.Read(promotionID)
	if err != nil {
		return fmt.Errorf("failed to read promotion: %w", err)
	}

	// Write the order-promotion relationship
	err = a.orderPromotionDAO.Write(orderID, promotionID)
	if err != nil {
		return fmt.Errorf("failed to apply promotion: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Applied promotion #%d to order #%d", promotionID, orderID))

	return nil
}

// GetOrderPromotions retrieves all promotions applied to an order
func (a *App) GetOrderPromotions(orderID uint64) ([]map[string]any, error) {
	orderPromotions, err := a.orderPromotionDAO.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(orderPromotions))
	for i, op := range orderPromotions {
		// Get promotion details
		promotion, err := a.promotionDAO.Read(op.PromotionID)
		if err != nil {
			// If promotion is deleted, still show the relationship with basic info
			result[i] = map[string]any{
				"id":   op.PromotionID,
				"name": "Deleted Promotion",
			}
			continue
		}

		result[i] = map[string]any{
			"id":         op.PromotionID,
			"name":       promotion.OwnerOrName,
			"totalPrice": promotion.TotalPrice,
			"itemCount":  promotion.ItemCount,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d promotions for order #%d", len(result), orderID))
	return result, nil
}

// GetPromotionOrders retrieves all orders that have a specific promotion applied
func (a *App) GetPromotionOrders(promotionID uint64) ([]map[string]any, error) {
	orderPromotions, err := a.orderPromotionDAO.GetByPromotionID(promotionID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(orderPromotions))
	for i, op := range orderPromotions {
		// Get order details
		order, err := a.orderDAO.Read(op.OrderID)
		if err != nil {
			// If order is deleted, still show the relationship with basic info
			result[i] = map[string]any{
				"orderID":      op.OrderID,
				"customerName": "Deleted Order",
			}
			continue
		}

		result[i] = map[string]any{
			"orderID":      op.OrderID,
			"customerName": order.OwnerOrName,
			"totalPrice":   order.TotalPrice,
			"itemCount":    order.ItemCount,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d orders for promotion #%d", len(result), promotionID))
	return result, nil
}

// RemovePromotionFromOrder removes a promotion from an order
func (a *App) RemovePromotionFromOrder(orderID, promotionID uint64) error {
	err := a.orderPromotionDAO.Delete(orderID, promotionID)
	if err != nil {
		return err
	}

	a.logger.Info(fmt.Sprintf("Removed promotion #%d from order #%d", promotionID, orderID))
	return nil
}

// GetOrderWithPromotions retrieves an order with all its promotions
func (a *App) GetOrderWithPromotions(orderID uint64) (map[string]any, error) {
	// Get order
	order, err := a.orderDAO.Read(orderID)
	if err != nil {
		return nil, err
	}

	// Get promotions
	promotions, err := a.GetOrderPromotions(orderID)
	if err != nil {
		return nil, err
	}

	// Calculate combined total price (items + promotions)
	combinedTotal := order.TotalPrice
	for _, promo := range promotions {
		if totalPrice, ok := promo["totalPrice"].(uint64); ok {
			combinedTotal += totalPrice
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved order #%d with %d promotions", orderID, len(promotions)))

	return map[string]any{
		"id":           order.ID,
		"customerName": order.OwnerOrName,
		"totalPrice":   combinedTotal,
		"promotions":   promotions,
		"itemCount":    order.ItemCount,
		"itemIDs":      order.ItemIDs,
	}, nil
}
