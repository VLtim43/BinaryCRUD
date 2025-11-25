package test

import (
	"BinaryCRUD/backend/dao"
	"fmt"
	"os"
	"testing"
)

// opTestCounter provides unique IDs for order_promotion test files
var opTestCounter uint64

// createOPTestFile creates a unique test file path and returns cleanup function
func createOPTestFile(prefix string) (string, func()) {
	opTestCounter++
	testFile := fmt.Sprintf("/tmp/%s_%d_%d.bin", prefix, os.Getpid(), opTestCounter)
	// Index is created in data/indexes/ by InitializeOrderPromotionIndex
	indexFile := fmt.Sprintf("data/indexes/%s_%d_%d.idx", prefix, os.Getpid(), opTestCounter)
	cleanup := func() {
		os.Remove(testFile)
		os.Remove(indexFile)
	}
	return testFile, cleanup
}

func TestOrderPromotionDAOWrite(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_write")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write first relationship
	err := opDAO.Write(1, 5)
	if err != nil {
		t.Fatalf("Failed to write order-promotion relationship: %v", err)
	}

	// Write second relationship
	err = opDAO.Write(2, 5)
	if err != nil {
		t.Fatalf("Failed to write second relationship: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Order-promotion file was not created")
	}
}

func TestOrderPromotionDAOPreventDuplicates(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_dup")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write relationship
	err := opDAO.Write(1, 5)
	if err != nil {
		t.Fatalf("Failed to write relationship: %v", err)
	}

	// Try to write same relationship again
	err = opDAO.Write(1, 5)
	if err == nil {
		t.Error("Expected error when writing duplicate relationship, got nil")
	}
}

func TestOrderPromotionDAOGetByOrderID(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_read")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write relationships
	_ = opDAO.Write(1, 5)
	_ = opDAO.Write(1, 7)
	_ = opDAO.Write(2, 5)

	// Get promotions for order 1
	promos, err := opDAO.GetByOrderID(1)
	if err != nil {
		t.Fatalf("Failed to get promotions for order 1: %v", err)
	}
	if len(promos) != 2 {
		t.Errorf("Expected 2 promotions for order 1, got %d", len(promos))
	}

	// Get promotions for non-existent order
	promos, err = opDAO.GetByOrderID(99)
	if err != nil {
		t.Fatalf("Failed to get promotions for non-existent order: %v", err)
	}
	if len(promos) != 0 {
		t.Errorf("Expected 0 promotions for order 99, got %d", len(promos))
	}
}

func TestOrderPromotionDAODelete(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_delete")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write and delete
	_ = opDAO.Write(1, 5)
	err := opDAO.Delete(1, 5)
	if err != nil {
		t.Fatalf("Failed to delete relationship: %v", err)
	}

	// Verify it's deleted by checking GetByOrderID
	promos, _ := opDAO.GetByOrderID(1)
	if len(promos) != 0 {
		t.Error("Relationship should be deleted")
	}
}

func TestOrderPromotionDAOGetOrderPromotions(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_get_order")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write multiple promotions for one order
	_ = opDAO.Write(1, 5)
	_ = opDAO.Write(1, 7)
	_ = opDAO.Write(1, 9)
	_ = opDAO.Write(2, 5) // Different order

	// Get promotions for order 1
	promos, err := opDAO.GetByOrderID(1)
	if err != nil {
		t.Fatalf("Failed to get order promotions: %v", err)
	}

	if len(promos) != 3 {
		t.Errorf("Expected 3 promotions for order 1, got %d", len(promos))
	}

	// Verify promotion IDs
	expectedPromos := map[uint64]bool{5: true, 7: true, 9: true}
	for _, promo := range promos {
		if !expectedPromos[promo.PromotionID] {
			t.Errorf("Unexpected promotion ID: %d", promo.PromotionID)
		}
	}
}

func TestOrderPromotionDAOGetPromotionOrders(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_get_promo")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write multiple orders for one promotion
	_ = opDAO.Write(1, 5)
	_ = opDAO.Write(2, 5)
	_ = opDAO.Write(3, 5)
	_ = opDAO.Write(1, 7) // Different promotion

	// Get orders for promotion 5
	orders, err := opDAO.GetByPromotionID(5)
	if err != nil {
		t.Fatalf("Failed to get promotion orders: %v", err)
	}

	if len(orders) != 3 {
		t.Errorf("Expected 3 orders for promotion 5, got %d", len(orders))
	}

	// Verify order IDs
	expectedOrders := map[uint64]bool{1: true, 2: true, 3: true}
	for _, order := range orders {
		if !expectedOrders[order.OrderID] {
			t.Errorf("Unexpected order ID: %d", order.OrderID)
		}
	}
}

func TestOrderPromotionDAOGetAll(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_get_all")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write multiple relationships
	_ = opDAO.Write(1, 5)
	_ = opDAO.Write(1, 7)
	_ = opDAO.Write(2, 5)

	// Get all relationships
	all, err := opDAO.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all relationships: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 relationships, got %d", len(all))
	}
}

func TestOrderPromotionDAODeletedRelationships(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_deleted")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Write and delete
	_ = opDAO.Write(1, 5)
	_ = opDAO.Write(1, 7)
	_ = opDAO.Delete(1, 5)

	// Get promotions for order 1 (should only get non-deleted)
	promos, err := opDAO.GetByOrderID(1)
	if err != nil {
		t.Fatalf("Failed to get order promotions: %v", err)
	}

	if len(promos) != 1 {
		t.Errorf("Expected 1 promotion after delete, got %d", len(promos))
	}

	if promos[0].PromotionID != 7 {
		t.Errorf("Expected promotion ID 7, got %d", promos[0].PromotionID)
	}
}

func TestOrderPromotionDAOEmptyResults(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_empty")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Get promotions for non-existent order
	promos, err := opDAO.GetByOrderID(999)
	if err != nil {
		t.Fatalf("Failed to get promotions: %v", err)
	}

	if len(promos) != 0 {
		t.Errorf("Expected 0 promotions, got %d", len(promos))
	}
}

func TestOrderPromotionDAOMultipleOrdersMultiplePromotions(t *testing.T) {
	testFile, cleanup := createOPTestFile("test_op_multiple")
	defer cleanup()

	opDAO := dao.NewOrderPromotionDAO(testFile)

	// Create a complex N:N relationship
	relationships := []struct {
		orderID     uint64
		promotionID uint64
	}{
		{1, 5}, {1, 7}, {1, 9},
		{2, 5}, {2, 9},
		{3, 7},
	}

	for _, rel := range relationships {
		err := opDAO.Write(rel.orderID, rel.promotionID)
		if err != nil {
			t.Fatalf("Failed to write relationship (%d, %d): %v", rel.orderID, rel.promotionID, err)
		}
	}

	// Verify order 1 has 3 promotions
	promos1, _ := opDAO.GetByOrderID(1)
	if len(promos1) != 3 {
		t.Errorf("Expected order 1 to have 3 promotions, got %d", len(promos1))
	}

	// Verify promotion 5 has 2 orders
	orders5, _ := opDAO.GetByPromotionID(5)
	if len(orders5) != 2 {
		t.Errorf("Expected promotion 5 to have 2 orders, got %d", len(orders5))
	}

	// Verify promotion 7 has 2 orders
	orders7, _ := opDAO.GetByPromotionID(7)
	if len(orders7) != 2 {
		t.Errorf("Expected promotion 7 to have 2 orders, got %d", len(orders7))
	}
}
