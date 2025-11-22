package test

import (
	"BinaryCRUD/backend/index"
	"fmt"
	"os"
	"testing"
)

func TestExtensibleHashBasicOperations(t *testing.T) {
	h := index.NewExtensibleHash(4)

	// Insert some entries
	err := h.Insert(1, 10, 100)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = h.Insert(2, 20, 200)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Search for existing entry
	offset, found := h.Search(1, 10)
	if !found {
		t.Error("Expected to find entry (1, 10)")
	}
	if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}

	// Search for non-existing entry
	_, found = h.Search(999, 999)
	if found {
		t.Error("Should not find non-existing entry")
	}
}

func TestExtensibleHashDuplicateKey(t *testing.T) {
	h := index.NewExtensibleHash(4)

	err := h.Insert(1, 10, 100)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Try to insert duplicate
	err = h.Insert(1, 10, 200)
	if err == nil {
		t.Error("Expected error for duplicate key")
	}
}

func TestExtensibleHashDelete(t *testing.T) {
	h := index.NewExtensibleHash(4)

	h.Insert(1, 10, 100)
	h.Insert(2, 20, 200)

	// Delete entry
	err := h.Delete(1, 10)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify deleted
	_, found := h.Search(1, 10)
	if found {
		t.Error("Entry should be deleted")
	}

	// Other entry should still exist
	offset, found := h.Search(2, 20)
	if !found {
		t.Error("Entry (2, 20) should still exist")
	}
	if offset != 200 {
		t.Errorf("Expected offset 200, got %d", offset)
	}
}

func TestExtensibleHashDeleteNotFound(t *testing.T) {
	h := index.NewExtensibleHash(4)

	err := h.Delete(999, 999)
	if err == nil {
		t.Error("Expected error for deleting non-existing key")
	}
}

func TestExtensibleHashBucketSplit(t *testing.T) {
	// Use small bucket size to force splits
	h := index.NewExtensibleHash(2)

	// Insert enough entries to cause splits
	for i := uint64(0); i < 10; i++ {
		err := h.Insert(i, i*10, int64(i*100))
		if err != nil {
			t.Fatalf("Failed to insert (%d, %d): %v", i, i*10, err)
		}
	}

	// Verify all entries are still searchable
	for i := uint64(0); i < 10; i++ {
		offset, found := h.Search(i, i*10)
		if !found {
			t.Errorf("Entry (%d, %d) not found after splits", i, i*10)
		}
		if offset != int64(i*100) {
			t.Errorf("Expected offset %d, got %d", i*100, offset)
		}
	}

	// Verify size
	if h.Size() != 10 {
		t.Errorf("Expected size 10, got %d", h.Size())
	}
}

func TestExtensibleHashGetByOrderID(t *testing.T) {
	h := index.NewExtensibleHash(4)

	// Order 1 has multiple promotions
	h.Insert(1, 10, 100)
	h.Insert(1, 20, 200)
	h.Insert(1, 30, 300)
	h.Insert(2, 10, 400)

	entries := h.GetByOrderID(1)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries for orderID=1, got %d", len(entries))
	}

	entries = h.GetByOrderID(2)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for orderID=2, got %d", len(entries))
	}
}

func TestExtensibleHashGetByPromotionID(t *testing.T) {
	h := index.NewExtensibleHash(4)

	// Promotion 10 is applied to multiple orders
	h.Insert(1, 10, 100)
	h.Insert(2, 10, 200)
	h.Insert(3, 10, 300)
	h.Insert(1, 20, 400)

	entries := h.GetByPromotionID(10)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries for promotionID=10, got %d", len(entries))
	}

	entries = h.GetByPromotionID(20)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for promotionID=20, got %d", len(entries))
	}
}

func TestExtensibleHashGetAll(t *testing.T) {
	h := index.NewExtensibleHash(4)

	h.Insert(1, 10, 100)
	h.Insert(2, 20, 200)
	h.Insert(3, 30, 300)

	entries := h.GetAll()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestExtensibleHashPersistence(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/test_ext_hash_%d.idx", os.Getpid())
	defer os.Remove(filePath)

	// Create and populate hash
	h := index.NewExtensibleHash(4)
	h.Insert(1, 10, 100)
	h.Insert(2, 20, 200)
	h.Insert(3, 30, 300)

	// Save to file
	err := h.Save(filePath)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load from file
	loaded, err := index.LoadExtensibleHash(filePath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify loaded data
	if loaded.Size() != 3 {
		t.Errorf("Expected size 3, got %d", loaded.Size())
	}

	offset, found := loaded.Search(1, 10)
	if !found {
		t.Error("Entry (1, 10) not found after load")
	}
	if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}

	offset, found = loaded.Search(2, 20)
	if !found {
		t.Error("Entry (2, 20) not found after load")
	}
	if offset != 200 {
		t.Errorf("Expected offset 200, got %d", offset)
	}
}

func TestExtensibleHashPersistenceAfterSplits(t *testing.T) {
	filePath := fmt.Sprintf("/tmp/test_ext_hash_split_%d.idx", os.Getpid())
	defer os.Remove(filePath)

	// Create with small bucket size to force splits
	h := index.NewExtensibleHash(2)

	// Insert many entries to cause splits
	for i := uint64(0); i < 20; i++ {
		h.Insert(i, i*10, int64(i*100))
	}

	originalDepth := h.GetGlobalDepth()
	originalSize := h.Size()

	// Save
	err := h.Save(filePath)
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load
	loaded, err := index.LoadExtensibleHash(filePath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify structure preserved
	if loaded.GetGlobalDepth() != originalDepth {
		t.Errorf("Expected global depth %d, got %d", originalDepth, loaded.GetGlobalDepth())
	}
	if loaded.Size() != originalSize {
		t.Errorf("Expected size %d, got %d", originalSize, loaded.Size())
	}

	// Verify all entries
	for i := uint64(0); i < 20; i++ {
		offset, found := loaded.Search(i, i*10)
		if !found {
			t.Errorf("Entry (%d, %d) not found after load", i, i*10)
		}
		if offset != int64(i*100) {
			t.Errorf("Expected offset %d, got %d", i*100, offset)
		}
	}
}

func TestExtensibleHashDirectoryDoubling(t *testing.T) {
	h := index.NewExtensibleHash(2)

	initialDirSize := h.GetDirectorySize()

	// Insert enough to cause directory doubling
	for i := uint64(0); i < 20; i++ {
		h.Insert(i, i, int64(i))
	}

	if h.GetDirectorySize() <= initialDirSize {
		t.Error("Directory should have grown")
	}

	// Verify depth increased
	if h.GetGlobalDepth() <= 1 {
		t.Error("Global depth should have increased")
	}
}

func TestExtensibleHashManyInserts(t *testing.T) {
	h := index.NewExtensibleHash(4)

	// Insert 1000 entries
	for i := uint64(0); i < 1000; i++ {
		err := h.Insert(i, i%100, int64(i))
		if err != nil {
			t.Fatalf("Failed to insert at %d: %v", i, err)
		}
	}

	if h.Size() != 1000 {
		t.Errorf("Expected size 1000, got %d", h.Size())
	}

	// Verify some random entries
	offset, found := h.Search(500, 0)
	if !found {
		t.Error("Entry (500, 0) not found")
	}
	if offset != 500 {
		t.Errorf("Expected offset 500, got %d", offset)
	}
}

func TestExtensibleHashSameOrderDifferentPromotions(t *testing.T) {
	h := index.NewExtensibleHash(4)

	// Same order with different promotions
	h.Insert(1, 1, 100)
	h.Insert(1, 2, 200)
	h.Insert(1, 3, 300)

	if h.Size() != 3 {
		t.Errorf("Expected size 3, got %d", h.Size())
	}

	// All should be found
	for i := uint64(1); i <= 3; i++ {
		offset, found := h.Search(1, i)
		if !found {
			t.Errorf("Entry (1, %d) not found", i)
		}
		if offset != int64(i*100) {
			t.Errorf("Expected offset %d, got %d", i*100, offset)
		}
	}
}
