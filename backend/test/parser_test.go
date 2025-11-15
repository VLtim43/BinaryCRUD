package test

import (
	"BinaryCRUD/backend/utils"
	"testing"
)

func TestParseItemEntry(t *testing.T) {
	// Create a valid item entry: [ID(2)][tombstone(1)][nameLength(2)][name...][price(4)]
	// ID=1, tombstone=0, name="Burger" (6 bytes), price=899
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	entryData = append(entryData, id...)

	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)

	// Name size (2 bytes)
	nameSize, _ := utils.WriteFixedNumber(2, 6)
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte("Burger")...)

	// Price (4 bytes)
	price, _ := utils.WriteFixedNumber(4, 899)
	entryData = append(entryData, price...)

	// Parse the entry
	item, err := utils.ParseItemEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse valid item entry: %v", err)
	}

	// Verify parsed values
	if item.ID != 1 {
		t.Errorf("Expected ID 1, got %d", item.ID)
	}
	if item.Name != "Burger" {
		t.Errorf("Expected name 'Burger', got '%s'", item.Name)
	}
	if item.Price != 899 {
		t.Errorf("Expected price 899, got %d", item.Price)
	}
	if item.Tombstone != 0x00 {
		t.Errorf("Expected tombstone 0x00, got 0x%02x", item.Tombstone)
	}
}

func TestParseItemEntryDeleted(t *testing.T) {
	// Create a deleted item entry (tombstone=1)
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 5)
	entryData = append(entryData, id...)

	// Tombstone (1 byte) - marked as deleted
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 1)
	entryData = append(entryData, tombstone...)

	// Name size (2 bytes)
	nameSize, _ := utils.WriteFixedNumber(2, 4)
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte("Soda")...)

	// Price (4 bytes)
	price, _ := utils.WriteFixedNumber(4, 199)
	entryData = append(entryData, price...)

	// Parse the entry
	item, err := utils.ParseItemEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse deleted item entry: %v", err)
	}

	// Verify tombstone is set
	if item.Tombstone != 0x01 {
		t.Errorf("Expected tombstone 0x01, got 0x%02x", item.Tombstone)
	}
	if item.ID != 5 {
		t.Errorf("Expected ID 5, got %d", item.ID)
	}
}

func TestParseItemEntryInvalid(t *testing.T) {
	// Test with entry too short
	shortEntry := []byte{0x00, 0x01}
	_, err := utils.ParseItemEntry(shortEntry)
	if err == nil {
		t.Error("Expected error for entry too short, got none")
	}

	// Test with missing name
	var incompleteEntry []byte
	id, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	incompleteEntry = append(incompleteEntry, id...)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	incompleteEntry = append(incompleteEntry, tombstone...)
	incompleteEntry = append(incompleteEntry, 0x1F)

	_, err = utils.ParseItemEntry(incompleteEntry)
	if err == nil {
		t.Error("Expected error for incomplete entry, got none")
	}
}

func TestParseCollectionEntry(t *testing.T) {
	// Create a valid collection entry
	// [ID(2)][tombstone(1)][nameLength(2)][name...][totalPrice(4)][itemCount(4)][itemIDs...]
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 10)
	entryData = append(entryData, id...)

	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)

	// Name size (2 bytes)
	nameSize, _ := utils.WriteFixedNumber(2, 4)
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte("John")...)

	// Total price (4 bytes)
	totalPrice, _ := utils.WriteFixedNumber(4, 1500)
	entryData = append(entryData, totalPrice...)

	// Item count (4 bytes)
	itemCount, _ := utils.WriteFixedNumber(4, 2)
	entryData = append(entryData, itemCount...)

	// Item IDs (2 bytes each)
	itemID1, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	entryData = append(entryData, itemID1...)
	itemID2, _ := utils.WriteFixedNumber(utils.IDSize, 3)
	entryData = append(entryData, itemID2...)

	// Parse the entry
	collection, err := utils.ParseCollectionEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse valid collection entry: %v", err)
	}

	// Verify parsed values
	if collection.ID != 10 {
		t.Errorf("Expected ID 10, got %d", collection.ID)
	}
	if collection.OwnerOrName != "John" {
		t.Errorf("Expected name 'John', got '%s'", collection.OwnerOrName)
	}
	if collection.TotalPrice != 1500 {
		t.Errorf("Expected total price 1500, got %d", collection.TotalPrice)
	}
	if collection.ItemCount != 2 {
		t.Errorf("Expected item count 2, got %d", collection.ItemCount)
	}
	if len(collection.ItemIDs) != 2 {
		t.Errorf("Expected 2 item IDs, got %d", len(collection.ItemIDs))
	}
	if collection.ItemIDs[0] != 1 {
		t.Errorf("Expected first item ID 1, got %d", collection.ItemIDs[0])
	}
	if collection.ItemIDs[1] != 3 {
		t.Errorf("Expected second item ID 3, got %d", collection.ItemIDs[1])
	}
	if collection.Tombstone != 0x00 {
		t.Errorf("Expected tombstone 0x00, got 0x%02x", collection.Tombstone)
	}
}

func TestParseCollectionEntryEmpty(t *testing.T) {
	// Create a collection with no items
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 20)
	entryData = append(entryData, id...)

	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)


	// Name size (2 bytes)
	nameSize, _ := utils.WriteFixedNumber(2, 4)
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte("Sale")...)


	// Total price (4 bytes)
	totalPrice, _ := utils.WriteFixedNumber(4, 0)
	entryData = append(entryData, totalPrice...)


	// Item count (4 bytes) - zero items
	itemCount, _ := utils.WriteFixedNumber(4, 0)
	entryData = append(entryData, itemCount...)


	// Parse the entry
	collection, err := utils.ParseCollectionEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse empty collection entry: %v", err)
	}

	// Verify parsed values
	if collection.ItemCount != 0 {
		t.Errorf("Expected item count 0, got %d", collection.ItemCount)
	}
	if len(collection.ItemIDs) != 0 {
		t.Errorf("Expected 0 item IDs, got %d", len(collection.ItemIDs))
	}
}

func TestParseCollectionEntryInvalid(t *testing.T) {
	// Test with entry too short
	shortEntry := []byte{0x00, 0x01}
	_, err := utils.ParseCollectionEntry(shortEntry)
	if err == nil {
		t.Error("Expected error for entry too short, got none")
	}

	// Test with truncated item IDs
	var truncatedEntry []byte
	id, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	truncatedEntry = append(truncatedEntry, id...)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	truncatedEntry = append(truncatedEntry, tombstone...)
	nameSize, _ := utils.WriteFixedNumber(2, 4)
	truncatedEntry = append(truncatedEntry, nameSize...)
	truncatedEntry = append(truncatedEntry, []byte("Test")...)
	totalPrice, _ := utils.WriteFixedNumber(4, 100)
	truncatedEntry = append(truncatedEntry, totalPrice...)
	itemCount, _ := utils.WriteFixedNumber(4, 5) // Says 5 items but we won't add them
	truncatedEntry = append(truncatedEntry, itemCount...)

	_, err = utils.ParseCollectionEntry(truncatedEntry)
	if err == nil {
		t.Error("Expected error for truncated item IDs, got none")
	}
}

func TestParseOrderPromotionEntry(t *testing.T) {
	// Create a valid order-promotion entry
	// Format: [orderID(2)][promotionID(2)][tombstone(1)]
	var entryData []byte

	// OrderID (2 bytes)
	orderID, _ := utils.WriteFixedNumber(utils.IDSize, 100)
	entryData = append(entryData, orderID...)


	// PromotionID (2 bytes)
	promotionID, _ := utils.WriteFixedNumber(utils.IDSize, 50)
	entryData = append(entryData, promotionID...)


	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)

	// Parse the entry
	orderPromo, err := utils.ParseOrderPromotionEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse valid order-promotion entry: %v", err)
	}

	// Verify parsed values
	if orderPromo.OrderID != 100 {
		t.Errorf("Expected order ID 100, got %d", orderPromo.OrderID)
	}
	if orderPromo.PromotionID != 50 {
		t.Errorf("Expected promotion ID 50, got %d", orderPromo.PromotionID)
	}
	if orderPromo.Tombstone != 0x00 {
		t.Errorf("Expected tombstone 0x00, got 0x%02x", orderPromo.Tombstone)
	}
}

func TestParseOrderPromotionEntryDeleted(t *testing.T) {
	// Create a deleted order-promotion entry
	var entryData []byte

	// OrderID (2 bytes)
	orderID, _ := utils.WriteFixedNumber(utils.IDSize, 200)
	entryData = append(entryData, orderID...)


	// PromotionID (2 bytes)
	promotionID, _ := utils.WriteFixedNumber(utils.IDSize, 75)
	entryData = append(entryData, promotionID...)


	// Tombstone (1 byte) - marked as deleted
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 1)
	entryData = append(entryData, tombstone...)

	// Parse the entry
	orderPromo, err := utils.ParseOrderPromotionEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse deleted order-promotion entry: %v", err)
	}

	// Verify tombstone is set
	if orderPromo.Tombstone != 0x01 {
		t.Errorf("Expected tombstone 0x01, got 0x%02x", orderPromo.Tombstone)
	}
}

func TestParseOrderPromotionEntryInvalid(t *testing.T) {
	// Test with entry too short
	shortEntry := []byte{0x00, 0x01}
	_, err := utils.ParseOrderPromotionEntry(shortEntry)
	if err == nil {
		t.Error("Expected error for entry too short, got none")
	}

	// Test with missing tombstone
	var incompleteEntry []byte
	orderID, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	incompleteEntry = append(incompleteEntry, orderID...)
	promotionID, _ := utils.WriteFixedNumber(utils.IDSize, 2)
	incompleteEntry = append(incompleteEntry, promotionID...)
	// Missing tombstone

	_, err = utils.ParseOrderPromotionEntry(incompleteEntry)
	if err == nil {
		t.Error("Expected error for missing tombstone, got none")
	}
}

func TestParseItemEntryShortName(t *testing.T) {
	// Create an item with a single character name
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 99)
	entryData = append(entryData, id...)

	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)


	// Name size (2 bytes) - 1 character
	nameSize, _ := utils.WriteFixedNumber(2, 1)
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte("X")...)


	// Price (4 bytes)
	price, _ := utils.WriteFixedNumber(4, 50)
	entryData = append(entryData, price...)

	// Parse the entry
	item, err := utils.ParseItemEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse item with short name: %v", err)
	}

	if item.Name != "X" {
		t.Errorf("Expected name 'X', got '%s'", item.Name)
	}
	if item.ID != 99 {
		t.Errorf("Expected ID 99, got %d", item.ID)
	}
	if item.Price != 50 {
		t.Errorf("Expected price 50, got %d", item.Price)
	}
}

func TestParseCollectionEntryLongName(t *testing.T) {
	// Create a collection with a long name
	longName := "ThisIsAVeryLongNameForACollectionThatShouldStillBeParsedCorrectly"
	var entryData []byte

	// ID (2 bytes)
	id, _ := utils.WriteFixedNumber(utils.IDSize, 77)
	entryData = append(entryData, id...)

	// Tombstone (1 byte)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)


	// Name size (2 bytes)
	nameSize, _ := utils.WriteFixedNumber(2, uint64(len(longName)))
	entryData = append(entryData, nameSize...)

	// Name
	entryData = append(entryData, []byte(longName)...)


	// Total price (4 bytes)
	totalPrice, _ := utils.WriteFixedNumber(4, 9999)
	entryData = append(entryData, totalPrice...)


	// Item count (4 bytes)
	itemCount, _ := utils.WriteFixedNumber(4, 0)
	entryData = append(entryData, itemCount...)


	// Parse the entry
	collection, err := utils.ParseCollectionEntry(entryData)
	if err != nil {
		t.Fatalf("Failed to parse collection with long name: %v", err)
	}

	if collection.OwnerOrName != longName {
		t.Errorf("Expected name '%s', got '%s'", longName, collection.OwnerOrName)
	}
}

// Benchmark tests
func BenchmarkParseItemEntry(b *testing.B) {
	// Create test data once
	var entryData []byte
	id, _ := utils.WriteFixedNumber(utils.IDSize, 1)
	entryData = append(entryData, id...)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)
	nameSize, _ := utils.WriteFixedNumber(2, 6)
	entryData = append(entryData, nameSize...)
	entryData = append(entryData, []byte("Burger")...)
	price, _ := utils.WriteFixedNumber(4, 899)
	entryData = append(entryData, price...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.ParseItemEntry(entryData)
	}
}

func BenchmarkParseCollectionEntry(b *testing.B) {
	// Create test data once
	var entryData []byte
	id, _ := utils.WriteFixedNumber(utils.IDSize, 10)
	entryData = append(entryData, id...)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)
	nameSize, _ := utils.WriteFixedNumber(2, 4)
	entryData = append(entryData, nameSize...)
	entryData = append(entryData, []byte("John")...)
	totalPrice, _ := utils.WriteFixedNumber(4, 1500)
	entryData = append(entryData, totalPrice...)
	itemCount, _ := utils.WriteFixedNumber(4, 3)
	entryData = append(entryData, itemCount...)
	for i := 0; i < 3; i++ {
		itemID, _ := utils.WriteFixedNumber(utils.IDSize, uint64(i))
		entryData = append(entryData, itemID...)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.ParseCollectionEntry(entryData)
	}
}

func BenchmarkParseOrderPromotionEntry(b *testing.B) {
	// Create test data once
	var entryData []byte
	orderID, _ := utils.WriteFixedNumber(utils.IDSize, 100)
	entryData = append(entryData, orderID...)
	promotionID, _ := utils.WriteFixedNumber(utils.IDSize, 50)
	entryData = append(entryData, promotionID...)
	tombstone, _ := utils.WriteFixedNumber(utils.TombstoneSize, 0)
	entryData = append(entryData, tombstone...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.ParseOrderPromotionEntry(entryData)
	}
}
