package test

import (
	"BinaryCRUD/backend/dao"
	"os"
	"testing"
)

func TestItemDAOWrite(t *testing.T) {
	testFile := "/tmp/test_item_write.bin"
	testIdx := "/tmp/test_item_write.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	// Create DAO
	itemDAO := dao.NewItemDAO(testFile)

	// Write first item
	err := itemDAO.Write("Burger", 899)
	if err != nil {
		t.Fatalf("Failed to write first item: %v", err)
	}

	// Write second item
	err = itemDAO.Write("Fries", 349)
	if err != nil {
		t.Fatalf("Failed to write second item: %v", err)
	}

	// Write third item
	err = itemDAO.Write("Soda", 199)
	if err != nil {
		t.Fatalf("Failed to write third item: %v", err)
	}

	// Verify index was created
	if _, err := os.Stat(testIdx); os.IsNotExist(err) {
		t.Error("Index file was not created")
	}

	// Verify index has 3 entries
	tree := itemDAO.GetIndexTree()
	if tree.Size() != 3 {
		t.Errorf("Expected 3 entries in index, got %d", tree.Size())
	}
}

func TestItemDAOReadWithIndex(t *testing.T) {
	testFile := "/tmp/test_item_read.bin"
	testIdx := "/tmp/test_item_read.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	// Create DAO and add items
	itemDAO := dao.NewItemDAO(testFile)
	itemDAO.Write("Pizza", 599)
	itemDAO.Write("Taco", 399)
	itemDAO.Write("Salad", 699)

	// Test indexed read
	id, name, price, err := itemDAO.ReadWithIndex(2, true)
	if err != nil {
		t.Fatalf("Failed to read item with index: %v", err)
	}

	if id != 2 {
		t.Errorf("Expected ID 2, got %d", id)
	}
	if name != "Taco" {
		t.Errorf("Expected name 'Taco', got '%s'", name)
	}
	if price != 399 {
		t.Errorf("Expected price 399, got %d", price)
	}

	// Test sequential read
	id, name, price, err = itemDAO.ReadWithIndex(1, false)
	if err != nil {
		t.Fatalf("Failed to read item sequentially: %v", err)
	}

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}
	if name != "Pizza" {
		t.Errorf("Expected name 'Pizza', got '%s'", name)
	}
	if price != 599 {
		t.Errorf("Expected price 599, got %d", price)
	}
}

func TestItemDAODelete(t *testing.T) {
	testFile := "/tmp/test_item_delete.bin"
	testIdx := "/tmp/test_item_delete.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	// Create DAO and add items
	itemDAO := dao.NewItemDAO(testFile)
	itemDAO.Write("Burger", 899)
	itemDAO.Write("Fries", 349)
	itemDAO.Write("Soda", 199)

	// Verify 3 items in index
	tree := itemDAO.GetIndexTree()
	if tree.Size() != 3 {
		t.Errorf("Expected 3 entries before delete, got %d", tree.Size())
	}

	// Delete item with ID 2
	err := itemDAO.Delete(2)
	if err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Verify index now has 2 entries (item 2 removed)
	if tree.Size() != 2 {
		t.Errorf("Expected 2 entries after delete, got %d", tree.Size())
	}

	// Verify item 2 is not in index
	_, found := tree.Search(2)
	if found {
		t.Error("Deleted item should not be in index")
	}

	// Verify reading deleted item fails
	_, _, _, err = itemDAO.ReadWithIndex(2, true)
	if err == nil {
		t.Error("Expected error when reading deleted item with index")
	}

	// Verify reading deleted item fails (sequential)
	_, _, _, err = itemDAO.ReadWithIndex(2, false)
	if err == nil {
		t.Error("Expected error when reading deleted item sequentially")
	}

	// Verify other items still readable
	id, name, price, err := itemDAO.ReadWithIndex(1, true)
	if err != nil {
		t.Fatalf("Failed to read non-deleted item: %v", err)
	}
	if id != 1 || name != "Burger" || price != 899 {
		t.Error("Non-deleted item data is incorrect")
	}
}

func TestItemDAOFullCRUDFlow(t *testing.T) {
	testFile := "/tmp/test_item_crud.bin"
	testIdx := "/tmp/test_item_crud.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	// Create DAO
	itemDAO := dao.NewItemDAO(testFile)

	// CREATE: Add 5 items
	items := []struct {
		name  string
		price uint64
	}{
		{"Burger", 899},
		{"Fries", 349},
		{"Soda", 199},
		{"Pizza", 599},
		{"Salad", 699},
	}

	for _, item := range items {
		err := itemDAO.Write(item.name, item.price)
		if err != nil {
			t.Fatalf("Failed to create item %s: %v", item.name, err)
		}
	}

	// Verify index has all 5 items
	tree := itemDAO.GetIndexTree()
	if tree.Size() != 5 {
		t.Errorf("Expected 5 items in index, got %d", tree.Size())
	}

	// READ: Verify all items can be read with index
	for i := uint64(1); i <= 5; i++ {
		id, name, price, err := itemDAO.ReadWithIndex(i, true)
		if err != nil {
			t.Fatalf("Failed to read item %d with index: %v", i, err)
		}

		expectedName := items[i-1].name
		expectedPrice := items[i-1].price

		if id != i {
			t.Errorf("Item %d: expected ID %d, got %d", i, i, id)
		}
		if name != expectedName {
			t.Errorf("Item %d: expected name '%s', got '%s'", i, expectedName, name)
		}
		if price != expectedPrice {
			t.Errorf("Item %d: expected price %d, got %d", i, expectedPrice, price)
		}
	}

	// READ: Verify all items can be read sequentially
	for i := uint64(1); i <= 5; i++ {
		_, _, _, err := itemDAO.ReadWithIndex(i, false)
		if err != nil {
			t.Fatalf("Failed to read item %d sequentially: %v", i, err)
		}
	}

	// DELETE: Delete items 2 and 4
	err := itemDAO.Delete(2)
	if err != nil {
		t.Fatalf("Failed to delete item 2: %v", err)
	}

	err = itemDAO.Delete(4)
	if err != nil {
		t.Fatalf("Failed to delete item 4: %v", err)
	}

	// Verify index now has 3 items
	if tree.Size() != 3 {
		t.Errorf("Expected 3 items after deletes, got %d", tree.Size())
	}

	// Verify deleted items return error
	_, _, _, err = itemDAO.ReadWithIndex(2, true)
	if err == nil {
		t.Error("Expected error reading deleted item 2")
	}

	_, _, _, err = itemDAO.ReadWithIndex(4, false)
	if err == nil {
		t.Error("Expected error reading deleted item 4")
	}

	// Verify non-deleted items still work
	remainingIDs := []uint64{1, 3, 5}
	for _, id := range remainingIDs {
		_, _, _, err := itemDAO.ReadWithIndex(id, true)
		if err != nil {
			t.Errorf("Failed to read non-deleted item %d: %v", id, err)
		}
	}

	// DELETE: Try to delete already deleted item
	err = itemDAO.Delete(2)
	if err == nil {
		t.Error("Expected error when deleting already deleted item")
	}

	// DELETE: Try to delete non-existent item
	err = itemDAO.Delete(99)
	if err == nil {
		t.Error("Expected error when deleting non-existent item")
	}
}

func TestItemDAOIndexPersistence(t *testing.T) {
	testFile := "/tmp/test_item_persist.bin"
	testIdx := "/tmp/test_item_persist.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	// Create DAO and add items
	itemDAO := dao.NewItemDAO(testFile)
	itemDAO.Write("Item1", 100)
	itemDAO.Write("Item2", 200)
	itemDAO.Write("Item3", 300)

	// Create NEW DAO instance (simulates app restart)
	itemDAO2 := dao.NewItemDAO(testFile)

	// Verify index was loaded from disk
	tree := itemDAO2.GetIndexTree()
	if tree.Size() != 3 {
		t.Errorf("Expected index to load 3 items from disk, got %d", tree.Size())
	}

	// Verify can read items with loaded index
	id, name, price, err := itemDAO2.ReadWithIndex(2, true)
	if err != nil {
		t.Fatalf("Failed to read with loaded index: %v", err)
	}

	if id != 2 || name != "Item2" || price != 200 {
		t.Error("Item data incorrect after loading index from disk")
	}
}

func TestItemDAOConcurrentWrites(t *testing.T) {
	testFile := "/tmp/test_item_concurrent.bin"
	testIdx := "/tmp/test_item_concurrent.idx"
	defer os.Remove(testFile)
	defer os.Remove(testIdx)

	itemDAO := dao.NewItemDAO(testFile)

	// Write items concurrently
	done := make(chan bool)

	write := func(name string, price uint64) {
		err := itemDAO.Write(name, price)
		if err != nil {
			t.Errorf("Concurrent write failed: %v", err)
		}
		done <- true
	}

	// Launch 5 concurrent writes
	go write("Item1", 100)
	go write("Item2", 200)
	go write("Item3", 300)
	go write("Item4", 400)
	go write("Item5", 500)

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all 5 items written
	tree := itemDAO.GetIndexTree()
	if tree.Size() != 5 {
		t.Errorf("Expected 5 items after concurrent writes, got %d", tree.Size())
	}

	// Verify all items are readable
	for i := uint64(1); i <= 5; i++ {
		_, _, _, err := itemDAO.ReadWithIndex(i, true)
		if err != nil {
			t.Errorf("Failed to read item %d after concurrent writes: %v", i, err)
		}
	}
}
