package main

import (
	"BinaryCRUD/backend/compression"
	"BinaryCRUD/backend/dao"
	"BinaryCRUD/backend/utils"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	toast             *Toast
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := NewLogger(1000) // Store up to 1000 log entries

	return &App{
		itemDAO:           dao.NewItemDAO("data/bin/items.bin"),
		orderDAO:          dao.NewOrderDAO("data/bin/orders.bin"),
		promotionDAO:      dao.NewPromotionDAO("data/bin/promotions.bin"),
		orderPromotionDAO: dao.NewOrderPromotionDAO("data/bin/order_promotions.bin"),
		logger:            logger,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.toast = NewToast(a)
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

// calculateTotalPriceWithValidation calculates total price and returns only valid items
// Returns (validItems, totalPrice) - items that exist and their total price
func (a *App) calculateTotalPriceWithValidation(itemIDs []uint64, entityName string) ([]uint64, uint64) {
	var totalPrice uint64
	var validItems []uint64

	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.Read(itemID)
		if err != nil {
			a.logger.Warn(fmt.Sprintf("Item ID %d in %s not found, skipping", itemID, entityName))
			continue
		}
		totalPrice += priceInCents
		validItems = append(validItems, itemID)
	}

	return validItems, totalPrice
}

// AddItem writes an item to the binary file with a price in cents and returns the assigned ID
func (a *App) AddItem(text string, priceInCents uint64) (uint64, error) {
	assignedID, err := a.itemDAO.Write(text, priceInCents)
	if err != nil {
		return 0, err
	}

	a.logger.Info(fmt.Sprintf("Created item #%d: %s ($%.2f)", assignedID, text, float64(priceInCents)/100))

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

// DeleteAllFiles deletes all generated data (bin, indexes, compressed) but keeps seed folder
func (a *App) DeleteAllFiles() error {
	// Use the shared cleanup utility
	results, err := utils.CleanupDataFiles(a.logger.Info)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("Error during cleanup: %v", err))
	}

	// Emit toast for each folder with deleted files
	folderNames := map[string]string{
		"data/bin":        "bin files",
		"data/indexes":    "indexes",
		"data/compressed": "compressed files",
	}

	totalDeleted := 0
	for _, result := range results {
		totalDeleted += result.Count
		if result.Count > 0 {
			name := folderNames[result.Folder]
			if name == "" {
				name = result.Folder
			}
			a.toast.Success(fmt.Sprintf("Deleted all %s (%d)", name, result.Count))
		}
	}

	if totalDeleted == 0 {
		a.toast.Info("No files to delete")
	}

	a.logger.Info(fmt.Sprintf("Deleted %d file(s) from bin, indexes, and compressed folders", totalDeleted))

	// Reload all DAOs to clear in-memory indexes
	a.itemDAO = dao.NewItemDAO("data/bin/items.bin")
	a.orderDAO = dao.NewOrderDAO("data/bin/orders.bin")
	a.promotionDAO = dao.NewPromotionDAO("data/bin/promotions.bin")
	a.orderPromotionDAO = dao.NewOrderPromotionDAO("data/bin/order_promotions.bin")
	a.logger.Info("Cleared all in-memory indexes")

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
	Owner        string   `json:"owner"`
	ItemIDs      []uint64 `json:"itemIDs"`
	PromotionIDs []uint64 `json:"promotionIDs,omitempty"`
}

// OrderPromotionEntry represents an order-promotion relationship in the JSON file
type OrderPromotionEntry struct {
	OrderID     uint64 `json:"orderID"`
	PromotionID uint64 `json:"promotionID"`
}

// IndexEntry represents an entry in a B+ tree index
type IndexEntry struct {
	ID     uint64 `json:"id"`
	Offset int64  `json:"offset"`
}

// getIndexContentsFromTree extracts and formats index contents from a B+ tree
func (a *App) getIndexContentsFromTree(allEntries map[uint64]int64, indexName string) map[string]any {
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

	a.logger.Info(fmt.Sprintf("%s index contains %d entries", indexName, len(entries)))

	return map[string]any{
		"count":   len(entries),
		"entries": entries,
	}
}

// GetIndexContents returns the contents of the item B+ tree index for debugging
func (a *App) GetIndexContents() (map[string]any, error) {
	tree := a.itemDAO.GetIndexTree()
	return a.getIndexContentsFromTree(tree.GetAll(), "Item"), nil
}

// GetOrderIndexContents returns the contents of the order B+ tree index for debugging
func (a *App) GetOrderIndexContents() (map[string]any, error) {
	tree := a.orderDAO.GetIndexTree()
	return a.getIndexContentsFromTree(tree.GetAll(), "Order"), nil
}

// GetPromotionIndexContents returns the contents of the promotion B+ tree index for debugging
func (a *App) GetPromotionIndexContents() (map[string]any, error) {
	tree := a.promotionDAO.GetIndexTree()
	return a.getIndexContentsFromTree(tree.GetAll(), "Promotion"), nil
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

	promoSuccessCount := 0
	promoFailCount := 0

	promoJsonPath := "data/seed/promotions.json"
	promoData, err := os.ReadFile(promoJsonPath)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("No promotions.json found, skipping promotions: %v", err))
	} else {
		// Parse JSON into slice of promotions
		var promotions []PromotionEntry
		err = json.Unmarshal(promoData, &promotions)
		if err != nil {
			return fmt.Errorf("failed to parse promotions.json: %w", err)
		}

		a.logger.Info(fmt.Sprintf("Starting promotion population with %d promotions", len(promotions)))

		for i, promo := range promotions {
			// Calculate total price for promotion
			totalPrice, err := a.calculateTotalPrice(promo.ItemIDs)
			if err != nil {
				a.logger.Warn(fmt.Sprintf("Failed to calculate price for promotion '%s': %v", promo.Name, err))
				// Use 0 if calculation fails
				totalPrice = 0
			}

			// Create promotion using CollectionDAO
			_, err = a.promotionDAO.Write(promo.Name, totalPrice, promo.ItemIDs)
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
	}

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

	// Track embedded promotions to apply after all orders are created
	type embeddedPromotion struct {
		orderID      uint64
		promotionIDs []uint64
	}
	var embeddedPromotions []embeddedPromotion

	for i, order := range orders {
		validItems, totalPrice := a.calculateTotalPriceWithValidation(order.ItemIDs, fmt.Sprintf("order '%s'", order.Owner))

		if len(validItems) == 0 {
			a.logger.Warn(fmt.Sprintf("Order %d (%s) has no valid items, skipping", i+1, order.Owner))
			orderFailCount++
			continue
		}

		orderID, err := a.orderDAO.Write(order.Owner, totalPrice, validItems)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add order %d (%s): %v", i+1, order.Owner, err))
			orderFailCount++
			continue
		}

		// Track embedded promotions for this order
		if len(order.PromotionIDs) > 0 {
			embeddedPromotions = append(embeddedPromotions, embeddedPromotion{
				orderID:      orderID,
				promotionIDs: order.PromotionIDs,
			})
		}

		orderSuccessCount++
		a.logger.Info(fmt.Sprintf("Added order %d/%d: %s with %d items ($%.2f)",
			i+1, len(orders), order.Owner, len(validItems), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Orders population complete: %d succeeded, %d failed", orderSuccessCount, orderFailCount))

	// Populate order-promotion relationships
	orderPromoJsonPath := "data/seed/order_promotions.json"
	orderPromoData, err := os.ReadFile(orderPromoJsonPath)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("No order_promotions.json found, skipping order-promotion relationships: %v", err))
	} else {
		var orderPromotions []OrderPromotionEntry
		err = json.Unmarshal(orderPromoData, &orderPromotions)
		if err != nil {
			return fmt.Errorf("failed to parse order_promotions.json: %w", err)
		}

		a.logger.Info(fmt.Sprintf("Starting order-promotion relationships with %d entries", len(orderPromotions)))

		orderPromoSuccessCount := 0
		orderPromoFailCount := 0

		for i, op := range orderPromotions {
			err := a.ApplyPromotionToOrder(op.OrderID, op.PromotionID)
			if err != nil {
				a.logger.Error(fmt.Sprintf("Failed to apply promotion %d to order %d: %v", op.PromotionID, op.OrderID, err))
				orderPromoFailCount++
				continue
			}

			orderPromoSuccessCount++
			a.logger.Info(fmt.Sprintf("Applied promotion #%d to order #%d (%d/%d)",
				op.PromotionID, op.OrderID, i+1, len(orderPromotions)))
		}

		a.logger.Info(fmt.Sprintf("Order-promotion relationships complete: %d succeeded, %d failed",
			orderPromoSuccessCount, orderPromoFailCount))
	}

	// Apply embedded promotions from orders.json
	if len(embeddedPromotions) > 0 {
		a.logger.Info(fmt.Sprintf("Applying %d embedded order-promotion relationships", len(embeddedPromotions)))

		embeddedSuccessCount := 0
		embeddedFailCount := 0

		for _, ep := range embeddedPromotions {
			for _, promoID := range ep.promotionIDs {
				err := a.ApplyPromotionToOrder(ep.orderID, promoID)
				if err != nil {
					a.logger.Error(fmt.Sprintf("Failed to apply embedded promotion %d to order %d: %v", promoID, ep.orderID, err))
					embeddedFailCount++
					continue
				}
				embeddedSuccessCount++
				a.logger.Info(fmt.Sprintf("Applied embedded promotion #%d to order #%d", promoID, ep.orderID))
			}
		}

		a.logger.Info(fmt.Sprintf("Embedded order-promotion relationships complete: %d succeeded, %d failed",
			embeddedSuccessCount, embeddedFailCount))
	}

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

// GetAllItems retrieves all items from the database, including deleted ones
func (a *App) GetAllItems() ([]map[string]any, error) {
	items, err := a.itemDAO.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"id":           item.ID,
			"name":         item.Name,
			"priceInCents": item.PriceInCents,
			"isDeleted":    item.IsDeleted,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d items", len(items)))
	return result, nil
}

// GetAllOrders retrieves all orders, including deleted ones
func (a *App) GetAllOrders() ([]map[string]any, error) {
	orders, err := a.orderDAO.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(orders))
	for i, order := range orders {
		result[i] = map[string]any{
			"id":           order.ID,
			"customerName": order.OwnerOrName,
			"totalPrice":   order.TotalPrice,
			"itemCount":    order.ItemCount,
			"itemIDs":      order.ItemIDs,
			"isDeleted":    order.IsDeleted,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d orders", len(orders)))
	return result, nil
}

// GetAllPromotions retrieves all promotions, including deleted ones
func (a *App) GetAllPromotions() ([]map[string]any, error) {
	promotions, err := a.promotionDAO.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(promotions))
	for i, promotion := range promotions {
		result[i] = map[string]any{
			"id":         promotion.ID,
			"name":       promotion.OwnerOrName,
			"totalPrice": promotion.TotalPrice,
			"itemCount":  promotion.ItemCount,
			"itemIDs":    promotion.ItemIDs,
			"isDeleted":  promotion.IsDeleted,
		}
	}

	a.logger.Info(fmt.Sprintf("Retrieved %d promotions", len(promotions)))
	return result, nil
}

// validateCollectionInput validates name and itemIDs for order/promotion creation
func (a *App) validateCollectionInput(name string, itemIDs []uint64, entityType string) error {
	if name == "" {
		return fmt.Errorf("%s name cannot be empty", entityType)
	}
	if len(itemIDs) == 0 {
		return fmt.Errorf("%s must contain at least one item", entityType)
	}
	return nil
}

// CreateOrder creates a new order with the given customer name and item IDs
func (a *App) CreateOrder(customerName string, itemIDs []uint64) (uint64, error) {
	if err := a.validateCollectionInput(customerName, itemIDs, "customer"); err != nil {
		return 0, err
	}

	totalPrice, err := a.calculateTotalPrice(itemIDs)
	if err != nil {
		return 0, err
	}

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
	if err := a.validateCollectionInput(promotionName, itemIDs, "promotion"); err != nil {
		return 0, err
	}

	totalPrice, err := a.calculateTotalPrice(itemIDs)
	if err != nil {
		return 0, err
	}

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

// CompressFile compresses a binary file using the specified algorithm
func (a *App) CompressFile(filename string, algorithm string) (map[string]any, error) {
	// Map filename to actual path (bin files are in data/bin/)
	inputPath := filepath.Join("data", "bin", filename)

	// Check if file exists
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	originalSize := fileInfo.Size()

	// Generate output filename (only huffman supported)
	outputFilename := filename + ".huffman.compressed"
	outputPath := filepath.Join("data", "compressed", outputFilename)

	hc := compression.NewHuffmanCompressor()
	err = hc.CompressFile(inputPath, outputPath)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Get compressed file size
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	// Delete original file and its index after successful compression
	utils.RemoveBinFile(filename, a.logger.Info)
	utils.RemoveIndexForBin(filename, a.logger.Info)

	// Calculate ratio
	ratio := float64(compressedSize) / float64(originalSize) * 100
	spaceSaved := float64(originalSize-compressedSize) / float64(originalSize) * 100

	a.logger.Info(fmt.Sprintf("Compressed %s -> %s (%.2f%% of original, saved %.2f%%)",
		filename, outputFilename, ratio, spaceSaved))

	return map[string]any{
		"outputFile":     outputFilename,
		"originalSize":   originalSize,
		"compressedSize": compressedSize,
		"ratio":          fmt.Sprintf("%.2f%%", ratio),
		"spaceSaved":     fmt.Sprintf("%.2f%%", spaceSaved),
	}, nil
}

// CompressAllFiles compresses all .bin files into a single archive
func (a *App) CompressAllFiles() (map[string]any, error) {
	binDir := filepath.Join("data", "bin")

	// Get all bin files
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bin directory: %w", err)
	}

	// Collect bin files
	var binFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".bin") {
			binFiles = append(binFiles, entry.Name())
		}
	}

	if len(binFiles) == 0 {
		return nil, fmt.Errorf("no .bin files found to compress")
	}

	// Build combined data with file markers
	// Format: [fileCount(4)][file1NameLen(2)][file1Name][file1Size(4)][file1Data]...
	var combined []byte

	// Write file count
	fileCountBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(fileCountBytes, uint32(len(binFiles)))
	combined = append(combined, fileCountBytes...)

	var totalOriginalSize int64

	for _, filename := range binFiles {
		filePath := filepath.Join(binDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}

		totalOriginalSize += int64(len(data))

		// Write filename length (2 bytes)
		nameBytes := []byte(filename)
		nameLenBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(nameLenBytes, uint16(len(nameBytes)))
		combined = append(combined, nameLenBytes...)

		// Write filename
		combined = append(combined, nameBytes...)

		// Write file size (4 bytes)
		sizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBytes, uint32(len(data)))
		combined = append(combined, sizeBytes...)

		// Write file data
		combined = append(combined, data...)
	}

	// Compress combined data
	hc := compression.NewHuffmanCompressor()
	compressedData, err := hc.Compress(combined)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Write to output file
	outputFilename := "all_files.huffman.compressed"
	outputPath := filepath.Join("data", "compressed", outputFilename)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, compressedData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write compressed file: %w", err)
	}

	compressedSize := int64(len(compressedData))

	// Delete original bin files and their indexes after successful compression
	for _, filename := range binFiles {
		utils.RemoveBinFile(filename, a.logger.Info)
		utils.RemoveIndexForBin(filename, a.logger.Info)
	}

	// Calculate ratio
	ratio := float64(compressedSize) / float64(totalOriginalSize) * 100
	spaceSaved := float64(totalOriginalSize-compressedSize) / float64(totalOriginalSize) * 100

	a.logger.Info(fmt.Sprintf("Compressed %d files -> %s (%.2f%% of original, saved %.2f%%)",
		len(binFiles), outputFilename, ratio, spaceSaved))

	return map[string]any{
		"outputFile":     outputFilename,
		"originalSize":   totalOriginalSize,
		"compressedSize": compressedSize,
		"ratio":          fmt.Sprintf("%.2f%%", ratio),
		"spaceSaved":     fmt.Sprintf("%.2f%%", spaceSaved),
	}, nil
}

// DecompressFile decompresses a compressed file
func (a *App) DecompressFile(filename string) (map[string]any, error) {
	inputPath := filepath.Join("data", "compressed", filename)

	// Check if file exists
	if _, err := os.Stat(inputPath); err != nil {
		return nil, fmt.Errorf("compressed file not found: %s", filename)
	}

	// Check if this is an all_files archive
	if filename == "all_files.huffman.compressed" {
		return a.decompressAllFiles(inputPath)
	}

	// Determine algorithm from filename
	if !strings.Contains(filename, ".huffman.") {
		return nil, fmt.Errorf("unknown compression format: %s", filename)
	}

	outputFilename := strings.TrimSuffix(filename, ".huffman.compressed")
	outputPath := filepath.Join("data", "bin", outputFilename)

	hc := compression.NewHuffmanCompressor()
	err := hc.DecompressFile(inputPath, outputPath)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// Get file sizes
	compressedInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	decompressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat decompressed file: %w", err)
	}
	originalSize := decompressedInfo.Size()

	// Delete compressed file after successful decompression
	utils.RemoveCompressedFile(filename, a.logger.Info)

	// Calculate ratio
	ratio := float64(compressedSize) / float64(originalSize) * 100
	spaceSaved := float64(originalSize-compressedSize) / float64(originalSize) * 100

	a.logger.Info(fmt.Sprintf("Decompressed %s -> %s (%d bytes)", filename, outputFilename, originalSize))

	return map[string]any{
		"outputFile":     outputFilename,
		"originalSize":   originalSize,
		"compressedSize": compressedSize,
		"ratio":          fmt.Sprintf("%.2f%%", ratio),
		"spaceSaved":     fmt.Sprintf("%.2f%%", spaceSaved),
	}, nil
}

// decompressAllFiles handles decompression of the all_files archive
func (a *App) decompressAllFiles(inputPath string) (map[string]any, error) {
	// Read compressed file
	compressedData, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed file: %w", err)
	}
	compressedSize := int64(len(compressedData))

	// Decompress
	hc := compression.NewHuffmanCompressor()
	data, err := hc.Decompress(compressedData)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// Parse the combined format
	// Format: [fileCount(4)][file1NameLen(2)][file1Name][file1Size(4)][file1Data]...
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid archive format: too short")
	}

	offset := 0

	// Read file count
	fileCount := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	binDir := filepath.Join("data", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bin directory: %w", err)
	}

	var totalOriginalSize int64
	var filesRestored int

	for i := uint32(0); i < fileCount; i++ {
		// Read filename length
		if offset+2 > len(data) {
			return nil, fmt.Errorf("invalid archive format: truncated at file %d name length", i)
		}
		nameLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		// Read filename
		if offset+int(nameLen) > len(data) {
			return nil, fmt.Errorf("invalid archive format: truncated at file %d name", i)
		}
		filename := string(data[offset : offset+int(nameLen)])
		offset += int(nameLen)

		// Read file size
		if offset+4 > len(data) {
			return nil, fmt.Errorf("invalid archive format: truncated at file %d size", i)
		}
		fileSize := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		// Read file data
		if offset+int(fileSize) > len(data) {
			return nil, fmt.Errorf("invalid archive format: truncated at file %d data", i)
		}
		fileData := data[offset : offset+int(fileSize)]
		offset += int(fileSize)

		// Write file
		filePath := filepath.Join(binDir, filename)
		if err := os.WriteFile(filePath, fileData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filename, err)
		}

		totalOriginalSize += int64(fileSize)
		filesRestored++
		a.logger.Info(fmt.Sprintf("Restored %s (%d bytes)", filename, fileSize))
	}

	// Delete compressed file after successful decompression
	utils.RemoveCompressedFile("all_files.huffman.compressed", a.logger.Info)

	// Calculate ratio
	ratio := float64(compressedSize) / float64(totalOriginalSize) * 100
	spaceSaved := float64(totalOriginalSize-compressedSize) / float64(totalOriginalSize) * 100

	a.logger.Info(fmt.Sprintf("Decompressed all_files archive: %d files restored (%d bytes total)", filesRestored, totalOriginalSize))

	return map[string]any{
		"outputFile":     fmt.Sprintf("%d files restored", filesRestored),
		"originalSize":   totalOriginalSize,
		"compressedSize": compressedSize,
		"ratio":          fmt.Sprintf("%.2f%%", ratio),
		"spaceSaved":     fmt.Sprintf("%.2f%%", spaceSaved),
	}, nil
}

// listFilesInDir is a helper to list files in a directory with optional filtering and mapping
func (a *App) listFilesInDir(dir string, filter func(string) bool, mapper func(string, int64) map[string]any) ([]map[string]any, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []map[string]any{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	files := make([]map[string]any, 0)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if filter != nil && !filter(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, mapper(entry.Name(), info.Size()))
	}

	return files, nil
}

// GetCompressedFiles returns a list of compressed files with metadata
func (a *App) GetCompressedFiles() ([]map[string]any, error) {
	compressedDir := filepath.Join("data", "compressed")

	if _, err := os.Stat(compressedDir); os.IsNotExist(err) {
		return []map[string]any{}, nil
	}

	entries, err := os.ReadDir(compressedDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed directory: %w", err)
	}

	files := make([]map[string]any, 0)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		compressedSize := info.Size()
		name := entry.Name()

		// Determine algorithm
		algorithm := "unknown"
		if strings.Contains(name, ".huffman.") {
			algorithm = "huffman"
		} else if strings.Contains(name, ".lzw.") {
			algorithm = "lzw"
		}

		// Read original size from file header (Huffman format: HUFF + uint32 originalSize)
		var originalSize int64 = 0
		filePath := filepath.Join(compressedDir, name)
		if algorithm == "huffman" {
			file, err := os.Open(filePath)
			if err == nil {
				header := make([]byte, 8) // 4 magic + 4 size
				if n, err := file.Read(header); err == nil && n == 8 {
					originalSize = int64(binary.LittleEndian.Uint32(header[4:8]))
				}
				file.Close()
			}
		}

		// Calculate ratio and space saved
		var ratio, spaceSaved float64
		if originalSize > 0 {
			ratio = float64(compressedSize) / float64(originalSize) * 100
			spaceSaved = float64(originalSize-compressedSize) / float64(originalSize) * 100
		}

		files = append(files, map[string]any{
			"name":           name,
			"originalSize":   originalSize,
			"compressedSize": compressedSize,
			"algorithm":      algorithm,
			"ratio":          fmt.Sprintf("%.2f%%", ratio),
			"spaceSaved":     fmt.Sprintf("%.2f%%", spaceSaved),
		})
	}

	return files, nil
}

// DeleteCompressedFile deletes a compressed file
func (a *App) DeleteCompressedFile(filename string) error {
	filePath := filepath.Join("data", "compressed", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Deleted compressed file: %s", filename))
	return nil
}

// GetBinFiles returns a list of .bin files in the data/bin directory
func (a *App) GetBinFiles() ([]map[string]any, error) {
	return a.listFilesInDir(
		filepath.Join("data", "bin"),
		func(name string) bool { return strings.HasSuffix(name, ".bin") },
		func(name string, size int64) map[string]any {
			return map[string]any{
				"name": name,
				"size": size,
			}
		},
	)
}
