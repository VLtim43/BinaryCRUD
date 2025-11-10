package test

import (
	"BinaryCRUD/backend/index"
	"os"
	"testing"
)

func TestBTreeBasicOperations(t *testing.T) {
	tree := index.NewBTree(4)

	// Test Insert
	err := tree.Insert(5, 100)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	err = tree.Insert(10, 200)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	err = tree.Insert(3, 50)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test duplicate
	err = tree.Insert(5, 150)
	if err == nil {
		t.Error("Expected error on duplicate insert")
	}

	// Test Search
	offset, found := tree.Search(5)
	if !found {
		t.Error("Expected to find ID 5")
	}
	if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}

	offset, found = tree.Search(10)
	if !found {
		t.Error("Expected to find ID 10")
	}
	if offset != 200 {
		t.Errorf("Expected offset 200, got %d", offset)
	}

	// Search non-existent
	_, found = tree.Search(99)
	if found {
		t.Error("Expected not to find ID 99")
	}

	// Test Size
	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	// Test Delete
	err = tree.Delete(5)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, found = tree.Search(5)
	if found {
		t.Error("Expected ID 5 to be deleted")
	}

	if tree.Size() != 2 {
		t.Errorf("Expected size 2 after delete, got %d", tree.Size())
	}
}

func TestBTreePersistence(t *testing.T) {
	tmpFile := "/tmp/test_btree.idx"
	defer os.Remove(tmpFile)

	// Create and populate tree
	tree := index.NewBTree(4)
	tree.Insert(1, 10)
	tree.Insert(2, 20)
	tree.Insert(3, 30)

	// Save
	err := tree.Save(tmpFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := index.Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if loaded.Size() != 3 {
		t.Errorf("Expected size 3, got %d", loaded.Size())
	}

	offset, found := loaded.Search(2)
	if !found || offset != 20 {
		t.Error("Failed to load tree correctly")
	}
}

func TestBTreeManyInserts(t *testing.T) {
	tree := index.NewBTree(4)

	// Insert many entries
	for i := uint64(1); i <= 100; i++ {
		err := tree.Insert(i, int64(i*10))
		if err != nil {
			t.Fatalf("Insert %d failed: %v", i, err)
		}
	}

	// Verify all
	for i := uint64(1); i <= 100; i++ {
		offset, found := tree.Search(i)
		if !found {
			t.Errorf("ID %d not found", i)
		}
		if offset != int64(i*10) {
			t.Errorf("ID %d: expected offset %d, got %d", i, i*10, offset)
		}
	}

	if tree.Size() != 100 {
		t.Errorf("Expected size 100, got %d", tree.Size())
	}
}
