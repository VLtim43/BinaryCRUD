package test

import (
	"BinaryCRUD/backend/utils"
	"os"
	"sync"
	"testing"
)

// Helper function to create a test file with items
func createTestFileWithItems(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header: 3 entities, 0 tombstones, next ID = 3
	header, _ := utils.WriteHeader("test.bin", 3, 0, 3)
	_, err = file.Write(header)
	if err != nil {
		return err
	}

	// Write 3 item entries
	for id := uint64(0); id < 3; id++ {
		var recordData []byte

		// ID (2 bytes)
		idBytes, _ := utils.WriteFixedNumber(utils.IDSize, id)
		recordData = append(recordData, idBytes...)

		// Tombstone (1 byte) - all active
		tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
		recordData = append(recordData, tombstone...)

		// Name size (2 bytes)
		name := "Item"
		nameSize, _ := utils.WriteFixedNumber(2, uint64(len(name)))
		recordData = append(recordData, nameSize...)

		// Name
		recordData = append(recordData, []byte(name)...)

		// Price (4 bytes)
		price, _ := utils.WriteFixedNumber(4, 100*(id+1))
		recordData = append(recordData, price...)

		// Record length prefix (2 bytes)
		recordLength, _ := utils.WriteFixedNumber(utils.RecordLengthSize, uint64(len(recordData)))

		// Write length prefix + record data
		_, err = file.Write(recordLength)
		if err != nil {
			return err
		}
		_, err = file.Write(recordData)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function to create a test file with order-promotion relationships
func createTestFileWithOrderPromotions(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header: 3 entities, 0 tombstones, next ID = 3
	header, _ := utils.WriteHeader("test.bin", 3, 0, 3)
	_, err = file.Write(header)
	if err != nil {
		return err
	}

	// Write 3 order-promotion entries
	relationships := []struct{ orderID, promotionID uint64 }{
		{1, 10},
		{1, 20},
		{2, 10},
	}

	for _, rel := range relationships {
		var recordData []byte

		// OrderID (2 bytes)
		orderID, _ := utils.WriteFixedNumber(utils.IDSize, rel.orderID)
		recordData = append(recordData, orderID...)

		// PromotionID (2 bytes)
		promotionID, _ := utils.WriteFixedNumber(utils.IDSize, rel.promotionID)
		recordData = append(recordData, promotionID...)

		// Tombstone (1 byte) - all active
		tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
		recordData = append(recordData, tombstone...)

		// Record length prefix (2 bytes)
		recordLength, _ := utils.WriteFixedNumber(utils.RecordLengthSize, uint64(len(recordData)))

		// Write length prefix + record data
		_, err = file.Write(recordLength)
		if err != nil {
			return err
		}
		_, err = file.Write(recordData)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestSoftDeleteByID(t *testing.T) {
	testFile := "/tmp/test_soft_delete.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete item with ID 1
	err = utils.SoftDeleteByID(testFile, 1, nil, nil)
	if err != nil {
		t.Fatalf("Failed to soft delete item: %v", err)
	}

	// Verify the tombstone was set
	entries, err := utils.SplitFileIntoEntries(testFile)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	// Find entry with ID 1
	found := false
	for _, entry := range entries {
		item, err := utils.ParseItemEntry(entry.Data)
		if err != nil {
			continue
		}

		if item.ID == 1 {
			found = true
			if item.Tombstone != 0x01 {
				t.Errorf("Expected tombstone 0x01 for deleted item, got 0x%02x", item.Tombstone)
			}
		}
	}

	if !found {
		t.Error("Could not find item with ID 1 after deletion")
	}

	// Verify header was updated
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, entitiesCount, tombstoneCount, nextID, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if entitiesCount != 3 {
		t.Errorf("Expected entities count 3, got %d", entitiesCount)
	}
	if tombstoneCount != 1 {
		t.Errorf("Expected tombstone count 1, got %d", tombstoneCount)
	}
	if nextID != 3 {
		t.Errorf("Expected next ID 3, got %d", nextID)
	}
}

func TestSoftDeleteByIDWithMutex(t *testing.T) {
	testFile := "/tmp/test_soft_delete_mutex.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mutex
	var mu sync.Mutex

	// Delete item with ID 0 with mutex
	err = utils.SoftDeleteByID(testFile, 0, &mu, nil)
	if err != nil {
		t.Fatalf("Failed to soft delete item with mutex: %v", err)
	}

	// Verify deletion worked
	entries, err := utils.SplitFileIntoEntries(testFile)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	for _, entry := range entries {
		item, err := utils.ParseItemEntry(entry.Data)
		if err != nil {
			continue
		}

		if item.ID == 0 {
			if item.Tombstone != 0x01 {
				t.Errorf("Expected tombstone 0x01 for deleted item, got 0x%02x", item.Tombstone)
			}
		}
	}
}

func TestSoftDeleteByIDWithIndexFunc(t *testing.T) {
	testFile := "/tmp/test_soft_delete_index.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Track if index delete function was called
	indexDeleteCalled := false
	var deletedID uint64

	indexDeleteFunc := func(id uint64) error {
		indexDeleteCalled = true
		deletedID = id
		return nil
	}

	// Delete item with ID 2 with index function
	err = utils.SoftDeleteByID(testFile, 2, nil, indexDeleteFunc)
	if err != nil {
		t.Fatalf("Failed to soft delete item: %v", err)
	}

	// Verify index delete function was called
	if !indexDeleteCalled {
		t.Error("Index delete function was not called")
	}
	if deletedID != 2 {
		t.Errorf("Expected deleted ID 2, got %d", deletedID)
	}
}

func TestSoftDeleteByIDNotFound(t *testing.T) {
	testFile := "/tmp/test_soft_delete_not_found.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items (IDs 0, 1, 2)
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to delete non-existent item with ID 99
	err = utils.SoftDeleteByID(testFile, 99, nil, nil)
	if err == nil {
		t.Error("Expected error when deleting non-existent item, got none")
	}

	// Verify header was not changed
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 0 {
		t.Errorf("Expected tombstone count 0 after failed delete, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByIDAlreadyDeleted(t *testing.T) {
	testFile := "/tmp/test_soft_delete_already_deleted.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete item with ID 1
	err = utils.SoftDeleteByID(testFile, 1, nil, nil)
	if err != nil {
		t.Fatalf("Failed to soft delete item: %v", err)
	}

	// Try to delete the same item again
	err = utils.SoftDeleteByID(testFile, 1, nil, nil)
	if err == nil {
		t.Error("Expected error when deleting already deleted item, got none")
	}

	// Verify tombstone count is still 1
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 1 {
		t.Errorf("Expected tombstone count 1 after duplicate delete attempt, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByCompositeKey(t *testing.T) {
	testFile := "/tmp/test_soft_delete_composite.bin"
	defer os.Remove(testFile)

	// Create test file with order-promotion relationships
	err := createTestFileWithOrderPromotions(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete relationship (1, 20)
	err = utils.SoftDeleteByCompositeKey(testFile, 1, 20, nil)
	if err != nil {
		t.Fatalf("Failed to soft delete by composite key: %v", err)
	}

	// Verify the tombstone was set
	entries, err := utils.SplitFileIntoEntries(testFile)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	// Find entry with composite key (1, 20)
	found := false
	for _, entry := range entries {
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			continue
		}

		if orderPromo.OrderID == 1 && orderPromo.PromotionID == 20 {
			found = true
			if orderPromo.Tombstone != 0x01 {
				t.Errorf("Expected tombstone 0x01 for deleted entry, got 0x%02x", orderPromo.Tombstone)
			}
		}
	}

	if !found {
		t.Error("Could not find entry with composite key (1, 20) after deletion")
	}

	// Verify header was updated
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 1 {
		t.Errorf("Expected tombstone count 1, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByCompositeKeyWithMutex(t *testing.T) {
	testFile := "/tmp/test_soft_delete_composite_mutex.bin"
	defer os.Remove(testFile)

	// Create test file with order-promotion relationships
	err := createTestFileWithOrderPromotions(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mutex
	var mu sync.Mutex

	// Delete relationship (2, 10) with mutex
	err = utils.SoftDeleteByCompositeKey(testFile, 2, 10, &mu)
	if err != nil {
		t.Fatalf("Failed to soft delete by composite key with mutex: %v", err)
	}

	// Verify deletion worked
	entries, err := utils.SplitFileIntoEntries(testFile)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	for _, entry := range entries {
		orderPromo, err := utils.ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			continue
		}

		if orderPromo.OrderID == 2 && orderPromo.PromotionID == 10 {
			if orderPromo.Tombstone != 0x01 {
				t.Errorf("Expected tombstone 0x01 for deleted entry, got 0x%02x", orderPromo.Tombstone)
			}
		}
	}
}

func TestSoftDeleteByCompositeKeyNotFound(t *testing.T) {
	testFile := "/tmp/test_soft_delete_composite_not_found.bin"
	defer os.Remove(testFile)

	// Create test file with order-promotion relationships
	err := createTestFileWithOrderPromotions(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to delete non-existent relationship (99, 99)
	err = utils.SoftDeleteByCompositeKey(testFile, 99, 99, nil)
	if err == nil {
		t.Error("Expected error when deleting non-existent composite key, got none")
	}

	// Verify header was not changed
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 0 {
		t.Errorf("Expected tombstone count 0 after failed delete, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByCompositeKeyAlreadyDeleted(t *testing.T) {
	testFile := "/tmp/test_soft_delete_composite_already_deleted.bin"
	defer os.Remove(testFile)

	// Create test file with order-promotion relationships
	err := createTestFileWithOrderPromotions(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete relationship (1, 10)
	err = utils.SoftDeleteByCompositeKey(testFile, 1, 10, nil)
	if err != nil {
		t.Fatalf("Failed to soft delete by composite key: %v", err)
	}

	// Try to delete the same relationship again
	err = utils.SoftDeleteByCompositeKey(testFile, 1, 10, nil)
	if err == nil {
		t.Error("Expected error when deleting already deleted composite key, got none")
	}

	// Verify tombstone count is still 1
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 1 {
		t.Errorf("Expected tombstone count 1 after duplicate delete attempt, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByIDMultipleDeletes(t *testing.T) {
	testFile := "/tmp/test_soft_delete_multiple.bin"
	defer os.Remove(testFile)

	// Create test file with 3 items
	err := createTestFileWithItems(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete all items
	for id := uint64(0); id < 3; id++ {
		err = utils.SoftDeleteByID(testFile, id, nil, nil)
		if err != nil {
			t.Fatalf("Failed to soft delete item %d: %v", id, err)
		}
	}

	// Verify all items are deleted
	entries, err := utils.SplitFileIntoEntries(testFile)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	for _, entry := range entries {
		item, err := utils.ParseItemEntry(entry.Data)
		if err != nil {
			continue
		}

		if item.Tombstone != 0x01 {
			t.Errorf("Expected all items to be deleted, but item %d has tombstone 0x%02x", item.ID, item.Tombstone)
		}
	}

	// Verify header shows 3 tombstones
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	_, _, tombstoneCount, _, err := utils.ReadHeader(file)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	if tombstoneCount != 3 {
		t.Errorf("Expected tombstone count 3, got %d", tombstoneCount)
	}
}

func TestSoftDeleteByIDFileNotExists(t *testing.T) {
	testFile := "/tmp/test_soft_delete_no_file.bin"

	// Try to delete from non-existent file
	err := utils.SoftDeleteByID(testFile, 1, nil, nil)
	if err == nil {
		t.Error("Expected error when deleting from non-existent file, got none")
	}
}

func TestSoftDeleteByCompositeKeyFileNotExists(t *testing.T) {
	testFile := "/tmp/test_soft_delete_composite_no_file.bin"

	// Try to delete from non-existent file
	err := utils.SoftDeleteByCompositeKey(testFile, 1, 1, nil)
	if err == nil {
		t.Error("Expected error when deleting from non-existent file, got none")
	}
}

// Benchmark tests
func BenchmarkSoftDeleteByID(b *testing.B) {
	testFile := "/tmp/bench_soft_delete.bin"
	defer os.Remove(testFile)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create fresh test file
		err := createTestFileWithItems(testFile)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		b.StartTimer()

		// Delete item with ID 1
		_ = utils.SoftDeleteByID(testFile, 1, nil, nil)
	}
}

func BenchmarkSoftDeleteByCompositeKey(b *testing.B) {
	testFile := "/tmp/bench_soft_delete_composite.bin"
	defer os.Remove(testFile)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create fresh test file
		err := createTestFileWithOrderPromotions(testFile)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		b.StartTimer()

		// Delete relationship (1, 20)
		_ = utils.SoftDeleteByCompositeKey(testFile, 1, 20, nil)
	}
}
