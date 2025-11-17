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

// Helper to create test app with minimal setup
func createTestApp() (*TestApp, func()) {
	itemFile := fmt.Sprintf("/tmp/test_app_items_%d.bin", os.Getpid())
	orderFile := fmt.Sprintf("/tmp/test_app_orders_%d.bin", os.Getpid())
	promoFile := fmt.Sprintf("/tmp/test_app_promos_%d.bin", os.Getpid())
	opFile := fmt.Sprintf("/tmp/test_app_op_%d.bin", os.Getpid())

	cleanup := func() {
		os.Remove(itemFile)
		os.Remove(orderFile)
		os.Remove(promoFile)
		os.Remove(opFile)
		os.Remove(itemFile[:len(itemFile)-4] + ".idx")
		os.Remove(orderFile[:len(orderFile)-4] + ".idx")
		os.Remove(promoFile[:len(promoFile)-4] + ".idx")
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

// calculateTotalPrice mimics the real implementation
func (a *TestApp) calculateTotalPrice(itemIDs []uint64) (uint64, error) {
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
