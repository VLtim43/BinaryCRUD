package test

import (
	"BinaryCRUD/backend/dao"
	"fmt"
	"os"
	"testing"
)

// Mock logger for testing
type TestLogger struct {
	logs []string
}

func (l *TestLogger) Info(msg string)  { l.logs = append(l.logs, "INFO: "+msg) }
func (l *TestLogger) Warn(msg string)  { l.logs = append(l.logs, "WARN: "+msg) }
func (l *TestLogger) Error(msg string) { l.logs = append(l.logs, "ERROR: "+msg) }

// testCounter provides unique IDs for test files
var testCounter uint64

// Helper to create test app with minimal setup
func createTestApp() (*TestApp, func()) {
	testCounter++
	uniqueID := fmt.Sprintf("%d_%d", os.Getpid(), testCounter)

	itemFile := fmt.Sprintf("/tmp/test_app_items_%s.bin", uniqueID)
	orderFile := fmt.Sprintf("/tmp/test_app_orders_%s.bin", uniqueID)
	promoFile := fmt.Sprintf("/tmp/test_app_promos_%s.bin", uniqueID)
	opFile := fmt.Sprintf("/tmp/test_app_op_%s.bin", uniqueID)

	cleanup := func() {
		// Remove data files from /tmp
		os.Remove(itemFile)
		os.Remove(orderFile)
		os.Remove(promoFile)
		os.Remove(opFile)
		// Remove index files from data/indexes/ (where InitializeDAOIndex creates them)
		os.Remove(fmt.Sprintf("data/indexes/test_app_items_%s.idx", uniqueID))
		os.Remove(fmt.Sprintf("data/indexes/test_app_orders_%s.idx", uniqueID))
		os.Remove(fmt.Sprintf("data/indexes/test_app_promos_%s.idx", uniqueID))
		os.Remove(fmt.Sprintf("data/indexes/test_app_op_%s.idx", uniqueID))
	}

	app := &TestApp{
		itemDAO:           dao.NewItemDAO(itemFile),
		orderDAO:          dao.NewOrderDAO(orderFile),
		promotionDAO:      dao.NewPromotionDAO(promoFile),
		orderPromotionDAO: dao.NewOrderPromotionDAO(opFile),
		logger:            &TestLogger{},
	}

	return app, cleanup
}

// Minimal App struct for testing (mimics real App)
type TestApp struct {
	itemDAO           *dao.ItemDAO
	orderDAO          *dao.OrderDAO
	promotionDAO      *dao.PromotionDAO
	orderPromotionDAO *dao.OrderPromotionDAO
	logger            *TestLogger
}

// AddItem writes an item to the binary file with a price in cents and returns the assigned ID
func (a *TestApp) AddItem(text string, priceInCents uint64) (uint64, error) {
	assignedID, err := a.itemDAO.Write(text, priceInCents)
	if err != nil {
		return 0, err
	}
	a.logger.Info(fmt.Sprintf("Created item #%d: %s ($%.2f)", assignedID, text, float64(priceInCents)/100))
	return assignedID, nil
}

// GetItem retrieves an item by ID from the binary file
func (a *TestApp) GetItem(id uint64) (map[string]any, error) {
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

// DeleteItem marks an item as deleted
func (a *TestApp) DeleteItem(id uint64) error {
	err := a.itemDAO.Delete(id)
	if err != nil {
		return err
	}
	a.logger.Info(fmt.Sprintf("Deleted item with ID: %d", id))
	return nil
}

// GetAllItems retrieves all items from the database
func (a *TestApp) GetAllItems() ([]map[string]any, error) {
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

// CreateOrder creates a new order with the given customer name and item IDs
func (a *TestApp) CreateOrder(customerName string, itemIDs []uint64) (uint64, error) {
	if customerName == "" {
		return 0, fmt.Errorf("customer name cannot be empty")
	}
	if len(itemIDs) == 0 {
		return 0, fmt.Errorf("order must contain at least one item")
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
func (a *TestApp) GetOrder(id uint64) (map[string]any, error) {
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
func (a *TestApp) DeleteOrder(id uint64) error {
	err := a.orderDAO.Delete(id)
	if err != nil {
		return err
	}
	a.logger.Info(fmt.Sprintf("Deleted order #%d", id))
	return nil
}

// GetAllOrders retrieves all orders
func (a *TestApp) GetAllOrders() ([]map[string]any, error) {
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

// CreatePromotion creates a new promotion with the given name and item IDs
func (a *TestApp) CreatePromotion(promotionName string, itemIDs []uint64) (uint64, error) {
	if promotionName == "" {
		return 0, fmt.Errorf("promotion name cannot be empty")
	}
	if len(itemIDs) == 0 {
		return 0, fmt.Errorf("promotion must contain at least one item")
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
func (a *TestApp) GetPromotion(id uint64) (map[string]any, error) {
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
func (a *TestApp) DeletePromotion(id uint64) error {
	err := a.promotionDAO.Delete(id)
	if err != nil {
		return err
	}
	a.logger.Info(fmt.Sprintf("Deleted promotion #%d", id))
	return nil
}

// GetAllPromotions retrieves all promotions
func (a *TestApp) GetAllPromotions() ([]map[string]any, error) {
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

// ApplyPromotionToOrder applies a promotion to an order
func (a *TestApp) ApplyPromotionToOrder(orderID, promotionID uint64) error {
	_, err := a.orderDAO.Read(orderID)
	if err != nil {
		return fmt.Errorf("failed to read order: %w", err)
	}
	_, err = a.promotionDAO.Read(promotionID)
	if err != nil {
		return fmt.Errorf("failed to read promotion: %w", err)
	}
	err = a.orderPromotionDAO.Write(orderID, promotionID)
	if err != nil {
		return fmt.Errorf("failed to apply promotion: %w", err)
	}
	a.logger.Info(fmt.Sprintf("Applied promotion #%d to order #%d", promotionID, orderID))
	return nil
}

// GetOrderPromotions retrieves all promotions applied to an order
func (a *TestApp) GetOrderPromotions(orderID uint64) ([]map[string]any, error) {
	orderPromotions, err := a.orderPromotionDAO.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, len(orderPromotions))
	for i, op := range orderPromotions {
		promotion, err := a.promotionDAO.Read(op.PromotionID)
		if err != nil {
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
func (a *TestApp) GetPromotionOrders(promotionID uint64) ([]map[string]any, error) {
	orderPromotions, err := a.orderPromotionDAO.GetByPromotionID(promotionID)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, len(orderPromotions))
	for i, op := range orderPromotions {
		order, err := a.orderDAO.Read(op.OrderID)
		if err != nil {
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
func (a *TestApp) RemovePromotionFromOrder(orderID, promotionID uint64) error {
	err := a.orderPromotionDAO.Delete(orderID, promotionID)
	if err != nil {
		return err
	}
	a.logger.Info(fmt.Sprintf("Removed promotion #%d from order #%d", promotionID, orderID))
	return nil
}

// GetOrderWithPromotions retrieves an order with all its promotions
func (a *TestApp) GetOrderWithPromotions(orderID uint64) (map[string]any, error) {
	order, err := a.orderDAO.Read(orderID)
	if err != nil {
		return nil, err
	}
	promotions, err := a.GetOrderPromotions(orderID)
	if err != nil {
		return nil, err
	}
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

// calculateTotalPrice mimics the real implementation
func (a *TestApp) calculateTotalPrice(itemIDs []uint64) (uint64, error) {
	var totalPrice uint64
	for _, itemID := range itemIDs {
		_, _, priceInCents, err := a.itemDAO.Read(itemID)
		if err != nil {
			return 0, fmt.Errorf("failed to read item %d: %w", itemID, err)
		}
		totalPrice += priceInCents
	}
	return totalPrice, nil
}

// calculateTotalPriceWithValidation mimics the real implementation
func (a *TestApp) calculateTotalPriceWithValidation(itemIDs []uint64, entityName string) ([]uint64, uint64) {
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

func TestCalculateTotalPrice(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test items
	_, _ = app.itemDAO.Write("Burger", 899)      // ID 0
	_, _ = app.itemDAO.Write("Fries", 349)       // ID 1
	_, _ = app.itemDAO.Write("Soda", 199)        // ID 2

	// Calculate total price
	total, err := app.calculateTotalPrice([]uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("Failed to calculate total price: %v", err)
	}

	expected := uint64(899 + 349 + 199)
	if total != expected {
		t.Errorf("Expected total %d, got %d", expected, total)
	}
}

func TestCalculateTotalPriceWithInvalidItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test items
	_, _ = app.itemDAO.Write("Burger", 899) // ID 0

	// Try to calculate with invalid item ID
	_, err := app.calculateTotalPrice([]uint64{0, 999})
	if err == nil {
		t.Error("Expected error for invalid item ID, got nil")
	}
}

func TestCalculateTotalPriceEmpty(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Calculate with no items
	total, err := app.calculateTotalPrice([]uint64{})
	if err != nil {
		t.Fatalf("Failed to calculate empty total: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0 for empty items, got %d", total)
	}
}

func TestCalculateTotalPriceWithValidation(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test items
	_, _ = app.itemDAO.Write("Burger", 899)  // ID 0
	_, _ = app.itemDAO.Write("Fries", 349)   // ID 1

	// Calculate with mix of valid and invalid items
	validItems, total := app.calculateTotalPriceWithValidation(
		[]uint64{0, 999, 1, 888},
		"test order",
	)

	// Should only include valid items (0, 1)
	if len(validItems) != 2 {
		t.Errorf("Expected 2 valid items, got %d", len(validItems))
	}

	expected := uint64(899 + 349)
	if total != expected {
		t.Errorf("Expected total %d, got %d", expected, total)
	}

	// Verify logger recorded warnings
	if len(app.logger.logs) != 2 {
		t.Errorf("Expected 2 warning logs, got %d", len(app.logger.logs))
	}
}

func TestCalculateTotalPriceWithValidationAllInvalid(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Don't add any items

	// Calculate with only invalid items
	validItems, total := app.calculateTotalPriceWithValidation(
		[]uint64{999, 888, 777},
		"test order",
	)

	if len(validItems) != 0 {
		t.Errorf("Expected 0 valid items, got %d", len(validItems))
	}

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
}

func TestCalculateTotalPriceWithDuplicateItems(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test item
	_, _ = app.itemDAO.Write("Burger", 899) // ID 0

	// Calculate with duplicate item IDs (simulating quantity)
	total, err := app.calculateTotalPrice([]uint64{0, 0, 0})
	if err != nil {
		t.Fatalf("Failed to calculate total: %v", err)
	}

	expected := uint64(899 * 3)
	if total != expected {
		t.Errorf("Expected total %d for 3 burgers, got %d", expected, total)
	}
}

func TestCalculateTotalPriceWithValidationPreservesOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test items
	_, _ = app.itemDAO.Write("Item1", 100) // ID 0
	_, _ = app.itemDAO.Write("Item2", 200) // ID 1
	_, _ = app.itemDAO.Write("Item3", 300) // ID 2

	// Calculate with some invalid items interspersed
	validItems, _ := app.calculateTotalPriceWithValidation(
		[]uint64{0, 999, 1, 888, 2},
		"test",
	)

	// Verify order is preserved
	expected := []uint64{0, 1, 2}
	for i, id := range validItems {
		if id != expected[i] {
			t.Errorf("Expected item %d at position %d, got %d", expected[i], i, id)
		}
	}
}

func TestCalculateTotalPriceLargeOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add test item
	_, _ = app.itemDAO.Write("Item", 100) // ID 0

	// Create large order (100 items)
	itemIDs := make([]uint64, 100)
	for i := range itemIDs {
		itemIDs[i] = 0
	}

	total, err := app.calculateTotalPrice(itemIDs)
	if err != nil {
		t.Fatalf("Failed to calculate large order: %v", err)
	}

	expected := uint64(100 * 100)
	if total != expected {
		t.Errorf("Expected total %d, got %d", expected, total)
	}
}

func TestCalculateTotalPriceZeroPriceItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add item with zero price
	_, _ = app.itemDAO.Write("Free Item", 0) // ID 0

	total, err := app.calculateTotalPrice([]uint64{0})
	if err != nil {
		t.Fatalf("Failed to calculate total: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0 for free item, got %d", total)
	}
}

// ==================== Item CRUD Tests ====================

func TestAddItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	id, err := app.AddItem("Test Item", 999)
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	if id != 0 {
		t.Errorf("Expected first item ID to be 0, got %d", id)
	}

	// Verify logger was called
	if len(app.logger.logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(app.logger.logs))
	}
}

func TestAddMultipleItems(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	id1, _ := app.AddItem("Item 1", 100)
	id2, _ := app.AddItem("Item 2", 200)
	id3, _ := app.AddItem("Item 3", 300)

	if id1 != 0 || id2 != 1 || id3 != 2 {
		t.Errorf("Expected sequential IDs 0,1,2 got %d,%d,%d", id1, id2, id3)
	}
}

func TestGetItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)

	item, err := app.GetItem(0)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if item["name"] != "Burger" {
		t.Errorf("Expected name 'Burger', got '%v'", item["name"])
	}
	if item["priceInCents"] != uint64(899) {
		t.Errorf("Expected price 899, got %v", item["priceInCents"])
	}
}

func TestGetItemNotFound(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.GetItem(999)
	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

func TestDeleteItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("ToDelete", 100)

	err := app.DeleteItem(0)
	if err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Item should not be readable after deletion
	_, err = app.GetItem(0)
	if err == nil {
		t.Error("Expected error reading deleted item")
	}
}

func TestDeleteItemNotFound(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	err := app.DeleteItem(999)
	if err == nil {
		t.Error("Expected error deleting non-existent item")
	}
}

func TestGetAllItems(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item 1", 100)
	_, _ = app.AddItem("Item 2", 200)
	_, _ = app.AddItem("Item 3", 300)

	items, err := app.GetAllItems()
	if err != nil {
		t.Fatalf("Failed to get all items: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestGetAllItemsEmpty(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	items, err := app.GetAllItems()
	if err != nil {
		t.Fatalf("Failed to get all items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

func TestGetAllItemsIncludesDeleted(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item 1", 100)
	_, _ = app.AddItem("Item 2", 200)
	_ = app.DeleteItem(0)

	items, err := app.GetAllItems()
	if err != nil {
		t.Fatalf("Failed to get all items: %v", err)
	}

	// Should include both items (deleted items are still in the list)
	if len(items) != 2 {
		t.Errorf("Expected 2 items (including deleted), got %d", len(items))
	}

	// Check that first item is marked as deleted
	if items[0]["isDeleted"] != true {
		t.Error("Expected first item to be marked as deleted")
	}
}

// ==================== Order CRUD Tests ====================

func TestCreateOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	// Add items first
	_, _ = app.AddItem("Burger", 899)
	_, _ = app.AddItem("Fries", 349)

	orderID, err := app.CreateOrder("John Doe", []uint64{0, 1})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if orderID != 0 {
		t.Errorf("Expected first order ID to be 0, got %d", orderID)
	}
}

func TestCreateOrderCalculatesTotalPrice(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)
	_, _ = app.AddItem("Fries", 349)

	_, _ = app.CreateOrder("Jane Doe", []uint64{0, 1})

	order, _ := app.GetOrder(0)
	expectedTotal := uint64(899 + 349)
	if order["totalPrice"] != expectedTotal {
		t.Errorf("Expected total price %d, got %v", expectedTotal, order["totalPrice"])
	}
}

func TestCreateOrderEmptyName(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)

	_, err := app.CreateOrder("", []uint64{0})
	if err == nil {
		t.Error("Expected error for empty customer name")
	}
}

func TestCreateOrderNoItems(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.CreateOrder("John", []uint64{})
	if err == nil {
		t.Error("Expected error for empty item list")
	}
}

func TestCreateOrderInvalidItem(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.CreateOrder("John", []uint64{999})
	if err == nil {
		t.Error("Expected error for invalid item ID")
	}
}

func TestGetOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)
	_, _ = app.CreateOrder("John Doe", []uint64{0})

	order, err := app.GetOrder(0)
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}

	if order["customer"] != "John Doe" {
		t.Errorf("Expected customer 'John Doe', got '%v'", order["customer"])
	}
}

func TestGetOrderNotFound(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.GetOrder(999)
	if err == nil {
		t.Error("Expected error for non-existent order")
	}
}

func TestDeleteOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})

	err := app.DeleteOrder(0)
	if err != nil {
		t.Fatalf("Failed to delete order: %v", err)
	}

	// Order should not be readable after deletion
	_, err = app.GetOrder(0)
	if err == nil {
		t.Error("Expected error reading deleted order")
	}
}

func TestGetAllOrders(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreateOrder("Jane", []uint64{0})

	orders, err := app.GetAllOrders()
	if err != nil {
		t.Fatalf("Failed to get all orders: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
}

// ==================== Promotion CRUD Tests ====================

func TestCreatePromotion(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)
	_, _ = app.AddItem("Fries", 349)

	promoID, err := app.CreatePromotion("Combo Deal", []uint64{0, 1})
	if err != nil {
		t.Fatalf("Failed to create promotion: %v", err)
	}

	if promoID != 0 {
		t.Errorf("Expected first promotion ID to be 0, got %d", promoID)
	}
}

func TestCreatePromotionEmptyName(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)

	_, err := app.CreatePromotion("", []uint64{0})
	if err == nil {
		t.Error("Expected error for empty promotion name")
	}
}

func TestCreatePromotionNoItems(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.CreatePromotion("Empty Promo", []uint64{})
	if err == nil {
		t.Error("Expected error for empty item list")
	}
}

func TestGetPromotion(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)
	_, _ = app.CreatePromotion("Burger Deal", []uint64{0})

	promo, err := app.GetPromotion(0)
	if err != nil {
		t.Fatalf("Failed to get promotion: %v", err)
	}

	if promo["name"] != "Burger Deal" {
		t.Errorf("Expected name 'Burger Deal', got '%v'", promo["name"])
	}
}

func TestGetPromotionNotFound(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.GetPromotion(999)
	if err == nil {
		t.Error("Expected error for non-existent promotion")
	}
}

func TestDeletePromotion(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreatePromotion("Promo", []uint64{0})

	err := app.DeletePromotion(0)
	if err != nil {
		t.Fatalf("Failed to delete promotion: %v", err)
	}

	_, err = app.GetPromotion(0)
	if err == nil {
		t.Error("Expected error reading deleted promotion")
	}
}

func TestGetAllPromotions(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreatePromotion("Promo 1", []uint64{0})
	_, _ = app.CreatePromotion("Promo 2", []uint64{0})

	promos, err := app.GetAllPromotions()
	if err != nil {
		t.Fatalf("Failed to get all promotions: %v", err)
	}

	if len(promos) != 2 {
		t.Errorf("Expected 2 promotions, got %d", len(promos))
	}
}

// ==================== Order-Promotion Relationship Tests ====================

func TestApplyPromotionToOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Discount", []uint64{0})

	err := app.ApplyPromotionToOrder(0, 0)
	if err != nil {
		t.Fatalf("Failed to apply promotion: %v", err)
	}
}

func TestApplyPromotionToOrderInvalidOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreatePromotion("Discount", []uint64{0})

	err := app.ApplyPromotionToOrder(999, 0)
	if err == nil {
		t.Error("Expected error for invalid order ID")
	}
}

func TestApplyPromotionToOrderInvalidPromotion(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})

	err := app.ApplyPromotionToOrder(0, 999)
	if err == nil {
		t.Error("Expected error for invalid promotion ID")
	}
}

func TestGetOrderPromotions(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Promo 1", []uint64{0})
	_, _ = app.CreatePromotion("Promo 2", []uint64{0})

	_ = app.ApplyPromotionToOrder(0, 0)
	_ = app.ApplyPromotionToOrder(0, 1)

	promos, err := app.GetOrderPromotions(0)
	if err != nil {
		t.Fatalf("Failed to get order promotions: %v", err)
	}

	if len(promos) != 2 {
		t.Errorf("Expected 2 promotions, got %d", len(promos))
	}
}

func TestGetOrderPromotionsEmpty(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})

	promos, err := app.GetOrderPromotions(0)
	if err != nil {
		t.Fatalf("Failed to get order promotions: %v", err)
	}

	if len(promos) != 0 {
		t.Errorf("Expected 0 promotions, got %d", len(promos))
	}
}

func TestGetPromotionOrders(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreateOrder("Jane", []uint64{0})
	_, _ = app.CreatePromotion("Discount", []uint64{0})

	_ = app.ApplyPromotionToOrder(0, 0)
	_ = app.ApplyPromotionToOrder(1, 0)

	orders, err := app.GetPromotionOrders(0)
	if err != nil {
		t.Fatalf("Failed to get promotion orders: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
}

func TestRemovePromotionFromOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Discount", []uint64{0})

	_ = app.ApplyPromotionToOrder(0, 0)

	err := app.RemovePromotionFromOrder(0, 0)
	if err != nil {
		t.Fatalf("Failed to remove promotion: %v", err)
	}

	promos, _ := app.GetOrderPromotions(0)
	if len(promos) != 0 {
		t.Errorf("Expected 0 promotions after removal, got %d", len(promos))
	}
}

func TestGetOrderWithPromotions(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Burger", 899)
	_, _ = app.AddItem("Fries", 349)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Side Deal", []uint64{1})

	_ = app.ApplyPromotionToOrder(0, 0)

	orderWithPromos, err := app.GetOrderWithPromotions(0)
	if err != nil {
		t.Fatalf("Failed to get order with promotions: %v", err)
	}

	// Total should be order items + promotion items
	expectedTotal := uint64(899 + 349) // Burger + Fries (from promotion)
	if orderWithPromos["totalPrice"] != expectedTotal {
		t.Errorf("Expected combined total %d, got %v", expectedTotal, orderWithPromos["totalPrice"])
	}

	promos := orderWithPromos["promotions"].([]map[string]any)
	if len(promos) != 1 {
		t.Errorf("Expected 1 promotion, got %d", len(promos))
	}
}

func TestGetOrderWithPromotionsNotFound(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, err := app.GetOrderWithPromotions(999)
	if err == nil {
		t.Error("Expected error for non-existent order")
	}
}

// ==================== Edge Case Tests ====================

func TestMultiplePromotionsOnOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item 1", 100)
	_, _ = app.AddItem("Item 2", 200)
	_, _ = app.AddItem("Item 3", 300)

	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Promo A", []uint64{1})
	_, _ = app.CreatePromotion("Promo B", []uint64{2})

	_ = app.ApplyPromotionToOrder(0, 0)
	_ = app.ApplyPromotionToOrder(0, 1)

	orderWithPromos, _ := app.GetOrderWithPromotions(0)

	// Total: Order(100) + PromoA(200) + PromoB(300) = 600
	expectedTotal := uint64(100 + 200 + 300)
	if orderWithPromos["totalPrice"] != expectedTotal {
		t.Errorf("Expected total %d, got %v", expectedTotal, orderWithPromos["totalPrice"])
	}
}

func TestOrderWithDeletedPromotion(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Promo", []uint64{0})

	_ = app.ApplyPromotionToOrder(0, 0)
	_ = app.DeletePromotion(0)

	promos, err := app.GetOrderPromotions(0)
	if err != nil {
		t.Fatalf("Failed to get promotions: %v", err)
	}

	// Should still have the relationship, but with "Deleted Promotion" name
	if len(promos) != 1 {
		t.Errorf("Expected 1 promotion entry, got %d", len(promos))
	}

	if promos[0]["name"] != "Deleted Promotion" {
		t.Errorf("Expected 'Deleted Promotion', got '%v'", promos[0]["name"])
	}
}

func TestPromotionWithDeletedOrder(t *testing.T) {
	app, cleanup := createTestApp()
	defer cleanup()

	_, _ = app.AddItem("Item", 100)
	_, _ = app.CreateOrder("John", []uint64{0})
	_, _ = app.CreatePromotion("Promo", []uint64{0})

	_ = app.ApplyPromotionToOrder(0, 0)
	_ = app.DeleteOrder(0)

	orders, err := app.GetPromotionOrders(0)
	if err != nil {
		t.Fatalf("Failed to get orders: %v", err)
	}

	// Should still have the relationship, but with "Deleted Order" name
	if len(orders) != 1 {
		t.Errorf("Expected 1 order entry, got %d", len(orders))
	}

	if orders[0]["customerName"] != "Deleted Order" {
		t.Errorf("Expected 'Deleted Order', got '%v'", orders[0]["customerName"])
	}
}
