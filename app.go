package main

import (
	"BinaryCRUD/backend/dao"
	"BinaryCRUD/backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
		itemDAO:           dao.NewItemDAO(utils.BinPath("items.bin")),
		orderDAO:          dao.NewOrderDAO(utils.BinPath("orders.bin")),
		promotionDAO:      dao.NewPromotionDAO(utils.BinPath("promotions.bin")),
		orderPromotionDAO: dao.NewOrderPromotionDAO(utils.BinPath("order_promotions.bin")),
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

// shutdown is called when the app is closing
// If CleanupOnExit flag is set to "true", it cleans up all data files
func (a *App) shutdown(ctx context.Context) {
	if CleanupOnExit == "true" {
		a.logger.Info("Application shutting down, cleaning up files...")
		a.cleanupOnExit()
		a.logger.Info("Cleanup complete, goodbye!")
	} else {
		a.logger.Info("Application shutting down, goodbye!")
	}
}

// cleanupOnExit deletes all data files silently (no toasts since UI is closing)
func (a *App) cleanupOnExit() {
	results, err := utils.CleanupDataFiles(a.logger.Info)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("Error during cleanup: %v", err))
	}

	totalDeleted := 0
	for _, result := range results {
		totalDeleted += result.Count
	}

	a.logger.Info(fmt.Sprintf("Shutdown cleanup: deleted %d file(s)", totalDeleted))
}

// PriceCalculationResult holds the result of a price calculation
type PriceCalculationResult struct {
	ValidItems []uint64
	TotalPrice uint64
	Errors     []error
}

// calculateTotalPrice calculates the total price of items by reading each item's price
// If strict is true, returns an error on the first missing item
// If strict is false, skips missing items and logs warnings
func (a *App) calculateTotalPrice(itemIDs []uint64, strict bool, entityName string) (*PriceCalculationResult, error) {
	result := &PriceCalculationResult{
		ValidItems: make([]uint64, 0, len(itemIDs)),
	}

	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.Read(itemID)
		if err != nil {
			if strict {
				return nil, fmt.Errorf("failed to read item %d: %w", itemID, err)
			}
			a.logger.Warn(fmt.Sprintf("Item ID %d in %s not found, skipping", itemID, entityName))
			result.Errors = append(result.Errors, err)
			continue
		}
		// Use safe addition to prevent overflow
		newTotal, err := utils.SafeAddUint64(result.TotalPrice, priceInCents)
		if err != nil {
			return nil, fmt.Errorf("price overflow calculating total for %s: %w", entityName, err)
		}
		result.TotalPrice = newTotal
		result.ValidItems = append(result.ValidItems, itemID)
	}

	return result, nil
}

// AddItem writes an item to the binary file with a price in cents and returns the assigned ID
func (a *App) AddItem(text string, priceInCents uint64) (uint64, error) {
	// Validate item name
	if err := utils.ValidateName(text); err != nil {
		return 0, fmt.Errorf("invalid item name: %w", err)
	}

	// Validate price
	if err := utils.ValidatePrice(priceInCents); err != nil {
		return 0, fmt.Errorf("invalid price: %w", err)
	}

	assignedID, err := a.itemDAO.Write(text, priceInCents)
	if err != nil {
		return 0, err
	}

	a.logger.Info(fmt.Sprintf("Created item #%d: %s ($%.2f)", assignedID, text, float64(priceInCents)/100))

	return assignedID, nil
}

// GetItem retrieves an item by ID from the binary file (uses index with automatic fallback)
func (a *App) GetItem(id uint64) (map[string]any, error) {
	itemID, name, priceInCents, err := a.itemDAO.Read(id)
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

// DeleteAllFiles deletes all generated data and keeps seed folder
func (a *App) DeleteAllFiles() error {
	results, err := utils.CleanupDataFiles(a.logger.Info)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("Error during cleanup: %v", err))
	}

	folderNames := map[string]string{
		utils.BinDir:   "bin files",
		utils.IndexDir: "indexes",
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

	a.logger.Info(fmt.Sprintf("Deleted %d file(s) from bin and indexes folders", totalDeleted))

	// Reload all DAOs to clear in-memory indexes
	a.itemDAO = dao.NewItemDAO(utils.BinPath("items.bin"))
	a.orderDAO = dao.NewOrderDAO(utils.BinPath("orders.bin"))
	a.promotionDAO = dao.NewPromotionDAO(utils.BinPath("promotions.bin"))
	a.orderPromotionDAO = dao.NewOrderPromotionDAO(utils.BinPath("order_promotions.bin"))
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

// populationResult tracks success/fail counts for a population operation
type populationResult struct {
	success int
	fail    int
}

// embeddedPromotion tracks order-promotion relationships from orders.json
type embeddedPromotion struct {
	orderID      uint64
	promotionIDs []uint64
}

// populateItems reads and populates items from seed file
func (a *App) populateItems() (*populationResult, error) {
	data, err := os.ReadFile(utils.SeedPath("items.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read items.json: %w", err)
	}

	var items []ItemEntry
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse items.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting data population with %d items", len(items)))
	result := &populationResult{}

	for i, item := range items {
		_, err := a.itemDAO.Write(item.Name, item.PriceInCents)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add item %d (%s): %v", i+1, item.Name, err))
			result.fail++
			continue
		}
		result.success++
		a.logger.Info(fmt.Sprintf("Added item %d/%d: %s ($%.2f)", i+1, len(items), item.Name, float64(item.PriceInCents)/100))
	}

	a.logger.Info(fmt.Sprintf("Items population complete: %d succeeded, %d failed", result.success, result.fail))
	return result, nil
}

// populatePromotions reads and populates promotions from seed file
func (a *App) populatePromotions() *populationResult {
	result := &populationResult{}

	data, err := os.ReadFile(utils.SeedPath("promotions.json"))
	if err != nil {
		a.logger.Warn(fmt.Sprintf("No promotions.json found, skipping promotions: %v", err))
		return result
	}

	var promotions []PromotionEntry
	if err := json.Unmarshal(data, &promotions); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to parse promotions.json: %v", err))
		return result
	}

	a.logger.Info(fmt.Sprintf("Starting promotion population with %d promotions", len(promotions)))

	for i, promo := range promotions {
		priceResult, err := a.calculateTotalPrice(promo.ItemIDs, false, fmt.Sprintf("promotion '%s'", promo.Name))
		totalPrice := uint64(0)
		if err == nil && priceResult != nil {
			totalPrice = priceResult.TotalPrice
		}

		_, err = a.promotionDAO.Write(promo.Name, totalPrice, promo.ItemIDs)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add promotion %d (%s): %v", i+1, promo.Name, err))
			result.fail++
			continue
		}
		result.success++
		a.logger.Info(fmt.Sprintf("Added promotion %d/%d: %s with %d items ($%.2f)",
			i+1, len(promotions), promo.Name, len(promo.ItemIDs), float64(totalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Promotions population complete: %d succeeded, %d failed", result.success, result.fail))
	return result
}

// populateOrders reads and populates orders from seed file, returns embedded promotions
func (a *App) populateOrders() (*populationResult, []embeddedPromotion, error) {
	data, err := os.ReadFile(utils.SeedPath("orders.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read orders.json: %w", err)
	}

	var orders []OrderEntry
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, nil, fmt.Errorf("failed to parse orders.json: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Starting orders population with %d orders", len(orders)))
	result := &populationResult{}
	var embedded []embeddedPromotion

	for i, order := range orders {
		priceResult, err := a.calculateTotalPrice(order.ItemIDs, false, fmt.Sprintf("order '%s'", order.Owner))
		if err != nil || priceResult == nil || len(priceResult.ValidItems) == 0 {
			a.logger.Warn(fmt.Sprintf("Order %d (%s) has no valid items, skipping", i+1, order.Owner))
			result.fail++
			continue
		}

		orderID, err := a.orderDAO.Write(order.Owner, priceResult.TotalPrice, priceResult.ValidItems)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Failed to add order %d (%s): %v", i+1, order.Owner, err))
			result.fail++
			continue
		}

		if len(order.PromotionIDs) > 0 {
			embedded = append(embedded, embeddedPromotion{orderID: orderID, promotionIDs: order.PromotionIDs})
		}

		result.success++
		a.logger.Info(fmt.Sprintf("Added order %d/%d: %s with %d items ($%.2f)",
			i+1, len(orders), order.Owner, len(priceResult.ValidItems), float64(priceResult.TotalPrice)/100))
	}

	a.logger.Info(fmt.Sprintf("Orders population complete: %d succeeded, %d failed", result.success, result.fail))
	return result, embedded, nil
}

// populateOrderPromotions reads and applies order-promotion relationships from seed file
func (a *App) populateOrderPromotions() *populationResult {
	result := &populationResult{}

	data, err := os.ReadFile(utils.SeedPath("order_promotions.json"))
	if err != nil {
		a.logger.Warn(fmt.Sprintf("No order_promotions.json found, skipping order-promotion relationships: %v", err))
		return result
	}

	var orderPromotions []OrderPromotionEntry
	if err := json.Unmarshal(data, &orderPromotions); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to parse order_promotions.json: %v", err))
		return result
	}

	a.logger.Info(fmt.Sprintf("Starting order-promotion relationships with %d entries", len(orderPromotions)))

	for i, op := range orderPromotions {
		if err := a.ApplyPromotionToOrder(op.OrderID, op.PromotionID); err != nil {
			a.logger.Error(fmt.Sprintf("Failed to apply promotion %d to order %d: %v", op.PromotionID, op.OrderID, err))
			result.fail++
			continue
		}
		result.success++
		a.logger.Info(fmt.Sprintf("Applied promotion #%d to order #%d (%d/%d)",
			op.PromotionID, op.OrderID, i+1, len(orderPromotions)))
	}

	a.logger.Info(fmt.Sprintf("Order-promotion relationships complete: %d succeeded, %d failed", result.success, result.fail))
	return result
}

// applyEmbeddedPromotions applies promotions embedded in orders.json
func (a *App) applyEmbeddedPromotions(embedded []embeddedPromotion) *populationResult {
	result := &populationResult{}
	if len(embedded) == 0 {
		return result
	}

	a.logger.Info(fmt.Sprintf("Applying %d embedded order-promotion relationships", len(embedded)))

	for _, ep := range embedded {
		for _, promoID := range ep.promotionIDs {
			if err := a.ApplyPromotionToOrder(ep.orderID, promoID); err != nil {
				a.logger.Error(fmt.Sprintf("Failed to apply embedded promotion %d to order %d: %v", promoID, ep.orderID, err))
				result.fail++
				continue
			}
			result.success++
			a.logger.Info(fmt.Sprintf("Applied embedded promotion #%d to order #%d", promoID, ep.orderID))
		}
	}

	a.logger.Info(fmt.Sprintf("Embedded order-promotion relationships complete: %d succeeded, %d failed", result.success, result.fail))
	return result
}

// PopulateInventory reads items and promotions from JSON files and adds them to the database
func (a *App) PopulateInventory() error {
	itemResult, err := a.populateItems()
	if err != nil {
		return err
	}
	a.toast.Success(fmt.Sprintf("Created items.bin (%d items)", itemResult.success))

	promoResult := a.populatePromotions()
	if promoResult.success > 0 {
		a.toast.Success(fmt.Sprintf("Created promotions.bin (%d promotions)", promoResult.success))
	}

	orderResult, embedded, err := a.populateOrders()
	if err != nil {
		return err
	}
	a.toast.Success(fmt.Sprintf("Created orders.bin (%d orders)", orderResult.success))

	opResult := a.populateOrderPromotions()
	embeddedResult := a.applyEmbeddedPromotions(embedded)
	totalOP := opResult.success + embeddedResult.success
	if totalOP > 0 {
		a.toast.Success(fmt.Sprintf("Created order_promotions.bin (%d relationships)", totalOP))
	}

	// Final summary
	totalSuccess := itemResult.success + promoResult.success + orderResult.success
	totalFail := itemResult.fail + promoResult.fail + orderResult.fail

	a.logger.Info(fmt.Sprintf("Total population complete: %d items + %d promotions + %d orders = %d total (%d failed)",
		itemResult.success, promoResult.success, orderResult.success, totalSuccess, totalFail))

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

// SearchItems searches for items by name.
func (a *App) SearchItems(pattern string) ([]map[string]any, error) {
	// First check if there are any items at all
	allItems, err := a.itemDAO.GetAll()
	if err != nil {
		return nil, err
	}
	if len(allItems) == 0 {
		return nil, fmt.Errorf("no items in database - populate inventory first")
	}

	items, err := a.itemDAO.SearchByName(pattern)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"id":           item.ID,
			"name":         item.Name,
			"priceInCents": item.PriceInCents,
		}
	}

	a.logger.Info(fmt.Sprintf("Search '%s' found %d items", pattern, len(items)))
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
			"id":         order.ID,
			"customer":   order.OwnerOrName,
			"totalPrice": order.TotalPrice,
			"itemCount":  order.ItemCount,
			"itemIDs":    order.ItemIDs,
			"isDeleted":  order.IsDeleted,
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
	if err := utils.ValidateName(name); err != nil {
		return fmt.Errorf("%s name: %w", entityType, err)
	}
	if err := utils.ValidateItemIDs(itemIDs); err != nil {
		return fmt.Errorf("%s: %w", entityType, err)
	}
	return nil
}

// CreateOrder creates a new order with the given customer name and item IDs
func (a *App) CreateOrder(customerName string, itemIDs []uint64) (uint64, error) {
	if err := a.validateCollectionInput(customerName, itemIDs, "customer"); err != nil {
		return 0, err
	}

	priceResult, err := a.calculateTotalPrice(itemIDs, true, "order")
	if err != nil {
		return 0, err
	}

	assignedID, err := a.orderDAO.Write(customerName, priceResult.TotalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Created order #%d for %s with %d items (total: $%.2f)",
		assignedID, customerName, len(itemIDs), float64(priceResult.TotalPrice)/100))

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
		"id":           order.ID,
		"customerName": order.OwnerOrName,
		"totalPrice":   order.TotalPrice,
		"itemCount":    order.ItemCount,
		"itemIDs":      order.ItemIDs,
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

	priceResult, err := a.calculateTotalPrice(itemIDs, true, "promotion")
	if err != nil {
		return 0, err
	}

	assignedID, err := a.promotionDAO.Write(promotionName, priceResult.TotalPrice, itemIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to create promotion: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Created promotion #%d: %s with %d items (total: $%.2f)",
		assignedID, promotionName, len(itemIDs), float64(priceResult.TotalPrice)/100))

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
				"orderID":  op.OrderID,
				"customer": "Deleted Order",
			}
			continue
		}

		result[i] = map[string]any{
			"orderID":    op.OrderID,
			"customer":   order.OwnerOrName,
			"totalPrice": order.TotalPrice,
			"itemCount":  order.ItemCount,
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

	// Calculate combined total price (items + promotions) with overflow checking
	combinedTotal := order.TotalPrice
	for _, promo := range promotions {
		if totalPrice, ok := promo["totalPrice"].(uint64); ok {
			newTotal, err := utils.SafeAddUint64(combinedTotal, totalPrice)
			if err != nil {
				return nil, fmt.Errorf("price overflow calculating combined total: %w", err)
			}
			combinedTotal = newTotal
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

// CompactResult represents the result of a compaction operation for frontend
type CompactResult struct {
	ItemsRemoved           int `json:"itemsRemoved"`
	OrdersAffected         int `json:"ordersAffected"`
	PromotionsAffected     int `json:"promotionsAffected"`
	OrdersRemoved          int `json:"ordersRemoved"`
	PromotionsRemoved      int `json:"promotionsRemoved"`
	OrderPromotionsRemoved int `json:"orderPromotionsRemoved"`
}

// Compact performs database compaction:
// - Removes all tombstoned (deleted) records from binary files
// - Updates orders/promotions to remove references to deleted items
// - Rebuilds all indexes
func (a *App) Compact() (*CompactResult, error) {
	a.logger.Info("Starting database compaction...")

	result, err := utils.CompactAll(
		utils.BinPath("items.bin"),
		utils.BinPath("orders.bin"),
		utils.BinPath("promotions.bin"),
		utils.BinPath("order_promotions.bin"),
	)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Compaction failed: %v", err))
		return nil, fmt.Errorf("compaction failed: %w", err)
	}

	// Reload all DAOs to rebuild indexes from the compacted files
	a.itemDAO = dao.NewItemDAO(utils.BinPath("items.bin"))
	a.orderDAO = dao.NewOrderDAO(utils.BinPath("orders.bin"))
	a.promotionDAO = dao.NewPromotionDAO(utils.BinPath("promotions.bin"))
	a.orderPromotionDAO = dao.NewOrderPromotionDAO(utils.BinPath("order_promotions.bin"))

	a.logger.Info("Indexes rebuilt after compaction")

	// Log summary
	a.logger.Info(fmt.Sprintf("Compaction complete: %d items removed, %d orders affected, %d promotions affected",
		result.ItemsRemoved, result.OrdersAffected, result.PromotionsAffected))

	// Show toast notifications
	if result.ItemsRemoved > 0 {
		a.toast.Success(fmt.Sprintf("Removed %d deleted items", result.ItemsRemoved))
	}
	if result.OrdersAffected > 0 {
		a.toast.Info(fmt.Sprintf("%d orders had item references cleaned", result.OrdersAffected))
	}
	if result.PromotionsAffected > 0 {
		a.toast.Info(fmt.Sprintf("%d promotions had item references cleaned", result.PromotionsAffected))
	}
	if result.OrdersRemoved > 0 {
		a.toast.Success(fmt.Sprintf("Removed %d deleted orders", result.OrdersRemoved))
	}
	if result.PromotionsRemoved > 0 {
		a.toast.Success(fmt.Sprintf("Removed %d deleted promotions", result.PromotionsRemoved))
	}
	if result.OrderPromotionsRemoved > 0 {
		a.toast.Success(fmt.Sprintf("Removed %d deleted order-promotion links", result.OrderPromotionsRemoved))
	}

	totalRemoved := result.ItemsRemoved + result.OrdersRemoved + result.PromotionsRemoved + result.OrderPromotionsRemoved
	totalAffected := result.OrdersAffected + result.PromotionsAffected

	if totalRemoved == 0 && totalAffected == 0 {
		a.toast.Info("No tombstoned records to compact")
	}

	return &CompactResult{
		ItemsRemoved:           result.ItemsRemoved,
		OrdersAffected:         result.OrdersAffected,
		PromotionsAffected:     result.PromotionsAffected,
		OrdersRemoved:          result.OrdersRemoved,
		PromotionsRemoved:      result.PromotionsRemoved,
		OrderPromotionsRemoved: result.OrderPromotionsRemoved,
	}, nil
}
