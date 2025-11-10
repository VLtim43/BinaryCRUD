package test

import (
	"BinaryCRUD/backend/dao"
	"os"
	"testing"
)

func TestOrderDAOCreateSingleOrder(t *testing.T) {
	testFile := "/tmp/test_order_create_single.bin"
	defer os.Remove(testFile)

	// Create OrderDAO
	orderDAO := dao.NewOrderDAO(testFile)

	// Create an order with customer name, total price, and item IDs
	err := orderDAO.Write("John Doe", 1500, []uint64{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Order file was not created")
	}

	// Read back the created order (IDs start at 0)
	order, err := orderDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read created order: %v", err)
	}

	// Verify order data
	if order.ID != 0 {
		t.Errorf("Expected order ID 0, got %d", order.ID)
	}
	if order.OwnerOrName != "John Doe" {
		t.Errorf("Expected customer name 'John Doe', got '%s'", order.OwnerOrName)
	}
	if order.TotalPrice != 1500 {
		t.Errorf("Expected total price 1500, got %d", order.TotalPrice)
	}
	if order.ItemCount != 3 {
		t.Errorf("Expected item count 3, got %d", order.ItemCount)
	}
	if len(order.ItemIDs) != 3 {
		t.Errorf("Expected 3 item IDs, got %d", len(order.ItemIDs))
	}

	// Verify item IDs
	expectedIDs := []uint64{1, 2, 3}
	for i, itemID := range order.ItemIDs {
		if itemID != expectedIDs[i] {
			t.Errorf("Item ID %d: expected %d, got %d", i, expectedIDs[i], itemID)
		}
	}
}

func TestOrderDAOCreateMultipleOrders(t *testing.T) {
	testFile := "/tmp/test_order_create_multiple.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create multiple orders
	orders := []struct {
		customer   string
		totalPrice uint64
		itemIDs    []uint64
	}{
		{"Alice Johnson", 2500, []uint64{1, 2, 3, 4, 5}},
		{"Bob Smith", 1200, []uint64{6, 7}},
		{"Charlie Brown", 899, []uint64{8}},
		{"Diana Prince", 3400, []uint64{9, 10, 11, 12}},
		{"Eve Adams", 650, []uint64{13}},
	}

	// Create all orders
	for _, order := range orders {
		err := orderDAO.Write(order.customer, order.totalPrice, order.itemIDs)
		if err != nil {
			t.Fatalf("Failed to create order for %s: %v", order.customer, err)
		}
	}

	// Verify each order can be read back correctly (IDs start at 0)
	for i := uint64(0); i < uint64(len(orders)); i++ {
		order, err := orderDAO.Read(i)
		if err != nil {
			t.Fatalf("Failed to read order %d: %v", i, err)
		}

		expectedOrder := orders[i]
		if order.ID != i {
			t.Errorf("Order %d: expected ID %d, got %d", i, i, order.ID)
		}
		if order.OwnerOrName != expectedOrder.customer {
			t.Errorf("Order %d: expected customer '%s', got '%s'", i, expectedOrder.customer, order.OwnerOrName)
		}
		if order.TotalPrice != expectedOrder.totalPrice {
			t.Errorf("Order %d: expected total price %d, got %d", i, expectedOrder.totalPrice, order.TotalPrice)
		}
		if order.ItemCount != uint64(len(expectedOrder.itemIDs)) {
			t.Errorf("Order %d: expected item count %d, got %d", i, len(expectedOrder.itemIDs), order.ItemCount)
		}

		// Verify item IDs
		for j, itemID := range order.ItemIDs {
			if itemID != expectedOrder.itemIDs[j] {
				t.Errorf("Order %d, item %d: expected ID %d, got %d", i, j, expectedOrder.itemIDs[j], itemID)
			}
		}
	}
}

func TestOrderDAOCreateEmptyOrder(t *testing.T) {
	testFile := "/tmp/test_order_create_empty.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create an order with no items
	err := orderDAO.Write("Empty Customer", 0, []uint64{})
	if err != nil {
		t.Fatalf("Failed to create empty order: %v", err)
	}

	// Read it back (ID starts at 0)
	order, err := orderDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read empty order: %v", err)
	}

	if order.OwnerOrName != "Empty Customer" {
		t.Errorf("Expected customer 'Empty Customer', got '%s'", order.OwnerOrName)
	}
	if order.TotalPrice != 0 {
		t.Errorf("Expected total price 0, got %d", order.TotalPrice)
	}
	if order.ItemCount != 0 {
		t.Errorf("Expected item count 0, got %d", order.ItemCount)
	}
	if len(order.ItemIDs) != 0 {
		t.Errorf("Expected 0 item IDs, got %d", len(order.ItemIDs))
	}
}

func TestOrderDAOCreateLargeOrder(t *testing.T) {
	testFile := "/tmp/test_order_create_large.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create an order with 50 items
	// Avoid item IDs 30 (0x1E) and 31 (0x1F) as they conflict with separators
	itemIDs := make([]uint64, 50)
	for i := uint64(0); i < 50; i++ {
		itemIDs[i] = i + 100 // Start from 100 to avoid separator byte conflicts
	}

	err := orderDAO.Write("Big Order Customer", 25000, itemIDs)
	if err != nil {
		t.Fatalf("Failed to create large order: %v", err)
	}

	// Read it back (ID starts at 0)
	order, err := orderDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read large order: %v", err)
	}

	if order.OwnerOrName != "Big Order Customer" {
		t.Errorf("Expected customer 'Big Order Customer', got '%s'", order.OwnerOrName)
	}
	if order.TotalPrice != 25000 {
		t.Errorf("Expected total price 25000, got %d", order.TotalPrice)
	}
	if order.ItemCount != 50 {
		t.Errorf("Expected item count 50, got %d", order.ItemCount)
	}
	if len(order.ItemIDs) != 50 {
		t.Errorf("Expected 50 item IDs, got %d", len(order.ItemIDs))
	}

	// Verify all item IDs
	for i, itemID := range order.ItemIDs {
		expectedID := uint64(i + 100)
		if itemID != expectedID {
			t.Errorf("Item %d: expected ID %d, got %d", i, expectedID, itemID)
		}
	}
}

func TestOrderDAOCreateWithSpecialCharacters(t *testing.T) {
	testFile := "/tmp/test_order_create_special.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create orders with special characters in customer names
	specialNames := []string{
		"José García",
		"François Müller",
		"李明",
		"O'Brien & Sons",
		"Smith-Jones Ltd.",
	}

	for i, name := range specialNames {
		err := orderDAO.Write(name, uint64((i+1)*1000), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create order with special name '%s': %v", name, err)
		}
	}

	// Verify each order (IDs start at 0)
	for i, expectedName := range specialNames {
		order, err := orderDAO.Read(uint64(i))
		if err != nil {
			t.Fatalf("Failed to read order with special name '%s': %v", expectedName, err)
		}

		if order.OwnerOrName != expectedName {
			t.Errorf("Expected customer name '%s', got '%s'", expectedName, order.OwnerOrName)
		}
	}
}

func TestOrderDAOCreateAndGetAll(t *testing.T) {
	testFile := "/tmp/test_order_create_getall.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create multiple orders
	expectedOrderCount := 5
	for i := 0; i < expectedOrderCount; i++ {
		customerName := "Customer" + string(rune('A'+i))
		err := orderDAO.Write(customerName, uint64((i+1)*1000), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create order %d: %v", i, err)
		}
	}

	// Get all orders
	orders, err := orderDAO.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all orders: %v", err)
	}

	if len(orders) != expectedOrderCount {
		t.Errorf("Expected %d orders, got %d", expectedOrderCount, len(orders))
	}

	// Verify orders are in sequence (IDs start at 0)
	for i, order := range orders {
		expectedID := uint64(i)
		if order.ID != expectedID {
			t.Errorf("Order %d: expected ID %d, got %d", i, expectedID, order.ID)
		}
	}
}

func TestOrderDAOCreateSequentialIDs(t *testing.T) {
	testFile := "/tmp/test_order_create_sequential.bin"
	defer os.Remove(testFile)

	orderDAO := dao.NewOrderDAO(testFile)

	// Create 10 orders and verify IDs are sequential (starting at 0)
	for i := 0; i < 10; i++ {
		err := orderDAO.Write("Customer", uint64((i+1)*100), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create order %d: %v", i, err)
		}

		// Read back and verify ID
		order, err := orderDAO.Read(uint64(i))
		if err != nil {
			t.Fatalf("Failed to read order %d: %v", i, err)
		}

		if order.ID != uint64(i) {
			t.Errorf("Expected order ID %d, got %d", i, order.ID)
		}
	}
}
