package test

import (
	"BinaryCRUD/backend/dao"
	"os"
	"testing"
)

func TestPromotionDAOCreateSinglePromotion(t *testing.T) {
	testFile := "/tmp/test_promotion_create_single.bin"
	defer os.Remove(testFile)

	// Create PromotionDAO
	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create a promotion with name, total price, and item IDs
	err := promotionDAO.Write("Summer Sale", 5000, []uint64{1, 2, 3, 4})
	if err != nil {
		t.Fatalf("Failed to create promotion: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Promotion file was not created")
	}

	// Read back the created promotion (IDs start at 0)
	promotion, err := promotionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read created promotion: %v", err)
	}

	// Verify promotion data
	if promotion.ID != 0 {
		t.Errorf("Expected promotion ID 0, got %d", promotion.ID)
	}
	if promotion.OwnerOrName != "Summer Sale" {
		t.Errorf("Expected promotion name 'Summer Sale', got '%s'", promotion.OwnerOrName)
	}
	if promotion.TotalPrice != 5000 {
		t.Errorf("Expected total price 5000, got %d", promotion.TotalPrice)
	}
	if promotion.ItemCount != 4 {
		t.Errorf("Expected item count 4, got %d", promotion.ItemCount)
	}
	if len(promotion.ItemIDs) != 4 {
		t.Errorf("Expected 4 item IDs, got %d", len(promotion.ItemIDs))
	}

	// Verify item IDs
	expectedIDs := []uint64{1, 2, 3, 4}
	for i, itemID := range promotion.ItemIDs {
		if itemID != expectedIDs[i] {
			t.Errorf("Item ID %d: expected %d, got %d", i, expectedIDs[i], itemID)
		}
	}
}

func TestPromotionDAOCreateMultiplePromotions(t *testing.T) {
	testFile := "/tmp/test_promotion_create_multiple.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create multiple promotions
	promotions := []struct {
		name       string
		totalPrice uint64
		itemIDs    []uint64
	}{
		{"Black Friday", 10000, []uint64{1, 2, 3, 4, 5, 6, 7}},
		{"Cyber Monday", 7500, []uint64{8, 9, 10}},
		{"New Year Sale", 6000, []uint64{11, 12, 13, 14}},
		{"Valentine's Special", 4500, []uint64{15, 16}},
		{"Easter Bundle", 3200, []uint64{17, 18, 19}},
	}

	// Create all promotions
	for _, promo := range promotions {
		err := promotionDAO.Write(promo.name, promo.totalPrice, promo.itemIDs)
		if err != nil {
			t.Fatalf("Failed to create promotion '%s': %v", promo.name, err)
		}
	}

	// Verify each promotion can be read back correctly (IDs start at 0)
	for i := uint64(0); i < uint64(len(promotions)); i++ {
		promotion, err := promotionDAO.Read(i)
		if err != nil {
			t.Fatalf("Failed to read promotion %d: %v", i, err)
		}

		expectedPromo := promotions[i]
		if promotion.ID != i {
			t.Errorf("Promotion %d: expected ID %d, got %d", i, i, promotion.ID)
		}
		if promotion.OwnerOrName != expectedPromo.name {
			t.Errorf("Promotion %d: expected name '%s', got '%s'", i, expectedPromo.name, promotion.OwnerOrName)
		}
		if promotion.TotalPrice != expectedPromo.totalPrice {
			t.Errorf("Promotion %d: expected total price %d, got %d", i, expectedPromo.totalPrice, promotion.TotalPrice)
		}
		if promotion.ItemCount != uint64(len(expectedPromo.itemIDs)) {
			t.Errorf("Promotion %d: expected item count %d, got %d", i, len(expectedPromo.itemIDs), promotion.ItemCount)
		}

		// Verify item IDs
		for j, itemID := range promotion.ItemIDs {
			if itemID != expectedPromo.itemIDs[j] {
				t.Errorf("Promotion %d, item %d: expected ID %d, got %d", i, j, expectedPromo.itemIDs[j], itemID)
			}
		}
	}
}

func TestPromotionDAOCreateEmptyPromotion(t *testing.T) {
	testFile := "/tmp/test_promotion_create_empty.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create a promotion with no items
	err := promotionDAO.Write("Empty Promotion", 0, []uint64{})
	if err != nil {
		t.Fatalf("Failed to create empty promotion: %v", err)
	}

	// Read it back (ID starts at 0)
	promotion, err := promotionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read empty promotion: %v", err)
	}

	if promotion.OwnerOrName != "Empty Promotion" {
		t.Errorf("Expected promotion name 'Empty Promotion', got '%s'", promotion.OwnerOrName)
	}
	if promotion.TotalPrice != 0 {
		t.Errorf("Expected total price 0, got %d", promotion.TotalPrice)
	}
	if promotion.ItemCount != 0 {
		t.Errorf("Expected item count 0, got %d", promotion.ItemCount)
	}
	if len(promotion.ItemIDs) != 0 {
		t.Errorf("Expected 0 item IDs, got %d", len(promotion.ItemIDs))
	}
}

func TestPromotionDAOCreateLargePromotion(t *testing.T) {
	testFile := "/tmp/test_promotion_create_large.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create a promotion with 100 items
	// Avoid item IDs 30 (0x1E) and 31 (0x1F) as they conflict with separators
	itemIDs := make([]uint64, 100)
	for i := uint64(0); i < 100; i++ {
		itemIDs[i] = i + 100 // Start from 100 to avoid separator byte conflicts
	}

	err := promotionDAO.Write("Mega Sale", 50000, itemIDs)
	if err != nil {
		t.Fatalf("Failed to create large promotion: %v", err)
	}

	// Read it back (ID starts at 0)
	promotion, err := promotionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read large promotion: %v", err)
	}

	if promotion.OwnerOrName != "Mega Sale" {
		t.Errorf("Expected promotion name 'Mega Sale', got '%s'", promotion.OwnerOrName)
	}
	if promotion.TotalPrice != 50000 {
		t.Errorf("Expected total price 50000, got %d", promotion.TotalPrice)
	}
	if promotion.ItemCount != 100 {
		t.Errorf("Expected item count 100, got %d", promotion.ItemCount)
	}
	if len(promotion.ItemIDs) != 100 {
		t.Errorf("Expected 100 item IDs, got %d", len(promotion.ItemIDs))
	}

	// Verify all item IDs
	for i, itemID := range promotion.ItemIDs {
		expectedID := uint64(i + 100)
		if itemID != expectedID {
			t.Errorf("Item %d: expected ID %d, got %d", i, expectedID, itemID)
		}
	}
}

func TestPromotionDAOCreateWithSpecialCharacters(t *testing.T) {
	testFile := "/tmp/test_promotion_create_special.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create promotions with special characters in names
	specialNames := []string{
		"50% OFF! 🎉",
		"Buy 1 Get 1 Free",
		"Spring-Summer 2024",
		"VIP Members' Exclusive",
		"Save $100 Today!",
	}

	for i, name := range specialNames {
		err := promotionDAO.Write(name, uint64((i+1)*2000), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create promotion with special name '%s': %v", name, err)
		}
	}

	// Verify each promotion (IDs start at 0)
	for i, expectedName := range specialNames {
		promotion, err := promotionDAO.Read(uint64(i))
		if err != nil {
			t.Fatalf("Failed to read promotion with special name '%s': %v", expectedName, err)
		}

		if promotion.OwnerOrName != expectedName {
			t.Errorf("Expected promotion name '%s', got '%s'", expectedName, promotion.OwnerOrName)
		}
	}
}

func TestPromotionDAOCreateAndGetAll(t *testing.T) {
	testFile := "/tmp/test_promotion_create_getall.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create multiple promotions
	promotionNames := []string{
		"Spring Sale",
		"Summer Special",
		"Fall Festival",
		"Winter Wonderland",
		"Year End Clearance",
	}

	for i, name := range promotionNames {
		err := promotionDAO.Write(name, uint64((i+1)*1500), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create promotion '%s': %v", name, err)
		}
	}

	// Get all promotions
	promotions, err := promotionDAO.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all promotions: %v", err)
	}

	if len(promotions) != len(promotionNames) {
		t.Errorf("Expected %d promotions, got %d", len(promotionNames), len(promotions))
	}

	// Verify promotions are in sequence (IDs start at 0)
	for i, promotion := range promotions {
		expectedID := uint64(i)
		if promotion.ID != expectedID {
			t.Errorf("Promotion %d: expected ID %d, got %d", i, expectedID, promotion.ID)
		}
		if promotion.OwnerOrName != promotionNames[i] {
			t.Errorf("Promotion %d: expected name '%s', got '%s'", i, promotionNames[i], promotion.OwnerOrName)
		}
	}
}

func TestPromotionDAOCreateSequentialIDs(t *testing.T) {
	testFile := "/tmp/test_promotion_create_sequential.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create 10 promotions and verify IDs are sequential (starting at 0)
	for i := 0; i < 10; i++ {
		promotionName := "Promotion " + string(rune('A'+i))
		err := promotionDAO.Write(promotionName, uint64((i+1)*500), []uint64{uint64(i)})
		if err != nil {
			t.Fatalf("Failed to create promotion %d: %v", i, err)
		}

		// Read back and verify ID
		promotion, err := promotionDAO.Read(uint64(i))
		if err != nil {
			t.Fatalf("Failed to read promotion %d: %v", i, err)
		}

		if promotion.ID != uint64(i) {
			t.Errorf("Expected promotion ID %d, got %d", i, promotion.ID)
		}
	}
}

func TestPromotionDAOCreateWithZeroPrice(t *testing.T) {
	testFile := "/tmp/test_promotion_create_zero.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create a free promotion (total price = 0)
	err := promotionDAO.Write("Free Sample Promotion", 0, []uint64{1, 2})
	if err != nil {
		t.Fatalf("Failed to create promotion with zero price: %v", err)
	}

	// Read it back (ID starts at 0)
	promotion, err := promotionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read promotion with zero price: %v", err)
	}

	if promotion.TotalPrice != 0 {
		t.Errorf("Expected total price 0, got %d", promotion.TotalPrice)
	}
	if promotion.ItemCount != 2 {
		t.Errorf("Expected item count 2, got %d", promotion.ItemCount)
	}
}

func TestPromotionDAOCreateWithHighPrice(t *testing.T) {
	testFile := "/tmp/test_promotion_create_high.bin"
	defer os.Remove(testFile)

	promotionDAO := dao.NewPromotionDAO(testFile)

	// Create a promotion with very high price
	highPrice := uint64(999999999)
	err := promotionDAO.Write("Premium Bundle", highPrice, []uint64{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to create promotion with high price: %v", err)
	}

	// Read it back (ID starts at 0)
	promotion, err := promotionDAO.Read(0)
	if err != nil {
		t.Fatalf("Failed to read promotion with high price: %v", err)
	}

	if promotion.TotalPrice != highPrice {
		t.Errorf("Expected total price %d, got %d", highPrice, promotion.TotalPrice)
	}
}
