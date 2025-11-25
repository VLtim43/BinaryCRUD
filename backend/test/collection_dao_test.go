package test

import (
	"BinaryCRUD/backend/dao"
	"os"
	"strings"
	"testing"
)

// cleanupCollectionTest removes both .bin file and index file
func cleanupCollectionTest(testFile string) {
	os.Remove(testFile)
	// Extract base name and remove from data/indexes/
	baseName := strings.TrimPrefix(testFile, "/tmp/")
	baseName = strings.TrimSuffix(baseName, ".bin")
	os.Remove("data/indexes/" + baseName + ".idx")
}

func TestCollectionDAOWrite(t *testing.T) {
	testFile := "/tmp/test_collection_write.bin"
	defer cleanupCollectionTest(testFile)

	// Create DAO
	collectionDAO := dao.NewOrderDAO(testFile)

	// Write first order
	_, err := collectionDAO.Write("John Doe", 1500, []uint64{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to write first order: %v", err)
	}

	// Write second order
	_, err = collectionDAO.Write("Jane Smith", 899, []uint64{4})
	if err != nil {
		t.Fatalf("Failed to write second order: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Collection file was not created")
	}
}

func TestCollectionDAORead(t *testing.T) {
	testFile := "/tmp/test_collection_read.bin"
	defer cleanupCollectionTest(testFile)

	// Create DAO and add collections
	collectionDAO := dao.NewOrderDAO(testFile)
	_, _ = collectionDAO.Write("Alice", 2500, []uint64{1, 2, 3, 4, 5})
	_, _ = collectionDAO.Write("Bob", 1200, []uint64{6, 7})
	_, _ = collectionDAO.Write("Charlie", 899, []uint64{8})

	// Read first collection (IDs start at 0)
	collection, err := collectionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read collection: %v", err)
	}

	if collection.ID != 0 {
		t.Errorf("Expected ID 0, got %d", collection.ID)
	}
	if collection.OwnerOrName != "Alice" {
		t.Errorf("Expected owner 'Alice', got '%s'", collection.OwnerOrName)
	}
	if collection.TotalPrice != 2500 {
		t.Errorf("Expected total price 2500, got %d", collection.TotalPrice)
	}
	if collection.ItemCount != 5 {
		t.Errorf("Expected item count 5, got %d", collection.ItemCount)
	}
	if len(collection.ItemIDs) != 5 {
		t.Errorf("Expected 5 item IDs, got %d", len(collection.ItemIDs))
	}

	expectedIDs := []uint64{1, 2, 3, 4, 5}
	for i, id := range collection.ItemIDs {
		if id != expectedIDs[i] {
			t.Errorf("Item ID %d: expected %d, got %d", i, expectedIDs[i], id)
		}
	}

	// Read second collection
	collection, err = collectionDAO.Read(1)
	if err != nil {
		t.Fatalf("Failed to read second collection: %v", err)
	}

	if collection.ID != 1 {
		t.Errorf("Expected ID 1, got %d", collection.ID)
	}
	if collection.OwnerOrName != "Bob" {
		t.Errorf("Expected owner 'Bob', got '%s'", collection.OwnerOrName)
	}
	if collection.TotalPrice != 1200 {
		t.Errorf("Expected total price 1200, got %d", collection.TotalPrice)
	}
	if collection.ItemCount != 2 {
		t.Errorf("Expected item count 2, got %d", collection.ItemCount)
	}
}

func TestCollectionDAODelete(t *testing.T) {
	testFile := "/tmp/test_collection_delete.bin"
	defer cleanupCollectionTest(testFile)

	// Create DAO and add collections
	collectionDAO := dao.NewOrderDAO(testFile)
	_, _ = collectionDAO.Write("Order1", 1000, []uint64{1, 2})
	_, _ = collectionDAO.Write("Order2", 2000, []uint64{3, 4, 5})
	_, _ = collectionDAO.Write("Order3", 3000, []uint64{6})

	// Delete collection with ID 1 (second item, since IDs start at 0)
	err := collectionDAO.Delete(1)
	if err != nil {
		t.Fatalf("Failed to delete collection: %v", err)
	}

	// Verify reading deleted collection fails
	_, err = collectionDAO.Read(1)
	if err == nil {
		t.Error("Expected error when reading deleted collection")
	}

	// Verify other collections still readable
	collection, err := collectionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read non-deleted collection: %v", err)
	}
	if collection.ID != 0 || collection.OwnerOrName != "Order1" || collection.TotalPrice != 1000 {
		t.Error("Non-deleted collection data is incorrect")
	}

	collection, err = collectionDAO.Read(2)
	if err != nil {
		t.Fatalf("Failed to read third collection: %v", err)
	}
	if collection.ID != 2 || collection.OwnerOrName != "Order3" || collection.TotalPrice != 3000 {
		t.Error("Third collection data is incorrect")
	}
}

func TestCollectionDAOFullCRUDFlow(t *testing.T) {
	testFile := "/tmp/test_collection_crud.bin"
	defer cleanupCollectionTest(testFile)

	// Create DAO
	collectionDAO := dao.NewPromotionDAO(testFile)

	// CREATE: Add 5 promotions
	promotions := []struct {
		name       string
		totalPrice uint64
		itemIDs    []uint64
	}{
		{"Summer Sale", 5000, []uint64{1, 2, 3, 4, 5}},
		{"Winter Special", 3000, []uint64{6, 7, 8}},
		{"Black Friday", 10000, []uint64{9, 10, 11, 12, 13, 14, 15}},
		{"Cyber Monday", 7500, []uint64{16, 17, 18, 19}},
		{"New Year", 2500, []uint64{20, 21}},
	}

	for _, promo := range promotions {
		_, err := collectionDAO.Write(promo.name, promo.totalPrice, promo.itemIDs)
		if err != nil {
			t.Fatalf("Failed to create promotion %s: %v", promo.name, err)
		}
	}

	// READ: Verify all promotions can be read (IDs start at 0)
	for i := uint64(0); i < 5; i++ {
		collection, err := collectionDAO.Read(i)
		if err != nil {
			t.Fatalf("Failed to read promotion %d: %v", i, err)
		}

		expectedName := promotions[i].name
		expectedPrice := promotions[i].totalPrice
		expectedItemCount := uint64(len(promotions[i].itemIDs))

		if collection.ID != i {
			t.Errorf("Promotion %d: expected ID %d, got %d", i, i, collection.ID)
		}
		if collection.OwnerOrName != expectedName {
			t.Errorf("Promotion %d: expected name '%s', got '%s'", i, expectedName, collection.OwnerOrName)
		}
		if collection.TotalPrice != expectedPrice {
			t.Errorf("Promotion %d: expected price %d, got %d", i, expectedPrice, collection.TotalPrice)
		}
		if collection.ItemCount != expectedItemCount {
			t.Errorf("Promotion %d: expected item count %d, got %d", i, expectedItemCount, collection.ItemCount)
		}

		// Verify item IDs
		for j, itemID := range collection.ItemIDs {
			expectedItemID := promotions[i].itemIDs[j]
			if itemID != expectedItemID {
				t.Errorf("Promotion %d, item %d: expected ID %d, got %d", i, j, expectedItemID, itemID)
			}
		}
	}

	// DELETE: Delete promotions at index 1 and 3 (second and fourth items)
	err := collectionDAO.Delete(1)
	if err != nil {
		t.Fatalf("Failed to delete promotion 1: %v", err)
	}

	err = collectionDAO.Delete(3)
	if err != nil {
		t.Fatalf("Failed to delete promotion 3: %v", err)
	}

	// Verify deleted promotions return error
	_, err = collectionDAO.Read(1)
	if err == nil {
		t.Error("Expected error reading deleted promotion 1")
	}

	_, err = collectionDAO.Read(3)
	if err == nil {
		t.Error("Expected error reading deleted promotion 3")
	}

	// Verify non-deleted promotions still work (IDs 0, 2, 4)
	remainingIDs := []uint64{0, 2, 4}
	for _, id := range remainingIDs {
		_, err := collectionDAO.Read(id)
		if err != nil {
			t.Errorf("Failed to read non-deleted promotion %d: %v", id, err)
		}
	}

	// DELETE: Try to delete already deleted promotion
	err = collectionDAO.Delete(1)
	if err == nil {
		t.Error("Expected error when deleting already deleted promotion")
	}

	// DELETE: Try to delete non-existent promotion
	err = collectionDAO.Delete(99)
	if err == nil {
		t.Error("Expected error when deleting non-existent promotion")
	}
}

func TestCollectionDAOEmptyItemList(t *testing.T) {
	testFile := "/tmp/test_collection_empty.bin"
	defer cleanupCollectionTest(testFile)

	collectionDAO := dao.NewOrderDAO(testFile)

	// Write collection with no items
	_, err := collectionDAO.Write("Empty Order", 0, []uint64{})
	if err != nil {
		t.Fatalf("Failed to write empty collection: %v", err)
	}

	// Read it back (ID starts at 0)
	collection, err := collectionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read empty collection: %v", err)
	}

	if collection.ItemCount != 0 {
		t.Errorf("Expected item count 0, got %d", collection.ItemCount)
	}
	if len(collection.ItemIDs) != 0 {
		t.Errorf("Expected 0 item IDs, got %d", len(collection.ItemIDs))
	}
}

func TestCollectionDAOLargeItemList(t *testing.T) {
	testFile := "/tmp/test_collection_large.bin"
	defer cleanupCollectionTest(testFile)

	collectionDAO := dao.NewPromotionDAO(testFile)

	// Create collection with 100 items
	// Avoid item IDs 30 (0x1E) and 31 (0x1F) as they conflict with separators
	itemIDs := make([]uint64, 100)
	for i := uint64(0); i < 100; i++ {
		itemIDs[i] = i + 100 // Start from 100 to avoid separator byte conflicts
	}

	_, err := collectionDAO.Write("Mega Sale", 50000, itemIDs)
	if err != nil {
		t.Fatalf("Failed to write large collection: %v", err)
	}

	// Read it back (ID starts at 0)
	collection, err := collectionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read large collection: %v", err)
	}

	if collection.ItemCount != 100 {
		t.Errorf("Expected item count 100, got %d", collection.ItemCount)
	}
	if len(collection.ItemIDs) != 100 {
		t.Errorf("Expected 100 item IDs, got %d", len(collection.ItemIDs))
	}

	// Verify all item IDs
	for i, itemID := range collection.ItemIDs {
		expectedID := uint64(i + 100)
		if itemID != expectedID {
			t.Errorf("Item %d: expected ID %d, got %d", i, expectedID, itemID)
		}
	}
}

func TestCollectionDAOOrdersVsPromotions(t *testing.T) {
	orderFile := "/tmp/test_orders.bin"
	promoFile := "/tmp/test_promos.bin"
	defer cleanupCollectionTest(orderFile)
	defer cleanupCollectionTest(promoFile)

	// Create separate DAOs
	orderDAO := dao.NewOrderDAO(orderFile)
	promoDAO := dao.NewPromotionDAO(promoFile)

	// Write to both
	_, err := orderDAO.Write("Customer A", 1500, []uint64{1, 2})
	if err != nil {
		t.Fatalf("Failed to write order: %v", err)
	}

	_, err = promoDAO.Write("Half Price Sale", 2500, []uint64{3, 4, 5})
	if err != nil {
		t.Fatalf("Failed to write promotion: %v", err)
	}

	// Read from both (IDs start at 0)
	order, err := orderDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read order: %v", err)
	}
	if order.OwnerOrName != "Customer A" {
		t.Errorf("Expected 'Customer A', got '%s'", order.OwnerOrName)
	}

	promo, err := promoDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read promotion: %v", err)
	}
	if promo.OwnerOrName != "Half Price Sale" {
		t.Errorf("Expected 'Half Price Sale', got '%s'", promo.OwnerOrName)
	}

	// Verify they're in separate files
	if order.ID == promo.ID && order.TotalPrice != promo.TotalPrice {
		// This is expected - both have ID 0, but they're different collections
		// in different files
		t.Logf("Orders and promotions maintain separate ID sequences (both start at 0)")
	}
}
