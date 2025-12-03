package test

import (
	"BinaryCRUD/backend/utils"
	"strings"
	"testing"
)

// ==================== Name Validation Tests ====================

func TestValidateNameEmpty(t *testing.T) {
	err := utils.ValidateName("")
	if err == nil {
		t.Error("Expected error for empty name")
	}
	if err != utils.ErrNameEmpty {
		t.Errorf("Expected ErrNameEmpty, got %v", err)
	}
}

func TestValidateNameValid(t *testing.T) {
	err := utils.ValidateName("Valid Name")
	if err != nil {
		t.Errorf("Expected no error for valid name, got %v", err)
	}
}

func TestValidateNameMaxLength(t *testing.T) {
	// Exactly at max length should be valid
	name := strings.Repeat("a", utils.MaxNameLength)
	err := utils.ValidateName(name)
	if err != nil {
		t.Errorf("Expected no error for max length name, got %v", err)
	}
}

func TestValidateNameTooLong(t *testing.T) {
	// One character over max length should fail
	name := strings.Repeat("a", utils.MaxNameLength+1)
	err := utils.ValidateName(name)
	if err == nil {
		t.Error("Expected error for name exceeding max length")
	}
	if err != utils.ErrNameTooLong {
		t.Errorf("Expected ErrNameTooLong, got %v", err)
	}
}

func TestValidateNameSpecialCharacters(t *testing.T) {
	// Special characters should be allowed
	specialNames := []string{
		"John's Order",
		"Café Special",
		"Item #123",
		"50% Off Deal",
		"Buy 1 Get 1 Free!",
		"日本語名前",
	}

	for _, name := range specialNames {
		err := utils.ValidateName(name)
		if err != nil {
			t.Errorf("Expected no error for special name '%s', got %v", name, err)
		}
	}
}

// ==================== Item IDs Validation Tests ====================

func TestValidateItemIDsEmpty(t *testing.T) {
	err := utils.ValidateItemIDs([]uint64{})
	if err == nil {
		t.Error("Expected error for empty item IDs")
	}
	if err != utils.ErrNoItems {
		t.Errorf("Expected ErrNoItems, got %v", err)
	}
}

func TestValidateItemIDsValid(t *testing.T) {
	err := utils.ValidateItemIDs([]uint64{1, 2, 3})
	if err != nil {
		t.Errorf("Expected no error for valid item IDs, got %v", err)
	}
}

func TestValidateItemIDsMaxCount(t *testing.T) {
	// Exactly at max should be valid
	itemIDs := make([]uint64, utils.MaxItemsPerCollection)
	for i := range itemIDs {
		itemIDs[i] = uint64(i)
	}
	err := utils.ValidateItemIDs(itemIDs)
	if err != nil {
		t.Errorf("Expected no error for max item count, got %v", err)
	}
}

func TestValidateItemIDsTooMany(t *testing.T) {
	// One over max should fail
	itemIDs := make([]uint64, utils.MaxItemsPerCollection+1)
	for i := range itemIDs {
		itemIDs[i] = uint64(i)
	}
	err := utils.ValidateItemIDs(itemIDs)
	if err == nil {
		t.Error("Expected error for exceeding max item count")
	}
	if err != utils.ErrTooManyItems {
		t.Errorf("Expected ErrTooManyItems, got %v", err)
	}
}

// ==================== Price Validation Tests ====================

func TestValidatePriceValid(t *testing.T) {
	err := utils.ValidatePrice(9999)
	if err != nil {
		t.Errorf("Expected no error for valid price, got %v", err)
	}
}

func TestValidatePriceZero(t *testing.T) {
	err := utils.ValidatePrice(0)
	if err != nil {
		t.Errorf("Expected no error for zero price, got %v", err)
	}
}

func TestValidatePriceMax(t *testing.T) {
	err := utils.ValidatePrice(utils.MaxPrice)
	if err != nil {
		t.Errorf("Expected no error for max price, got %v", err)
	}
}

func TestValidatePriceTooHigh(t *testing.T) {
	err := utils.ValidatePrice(utils.MaxPrice + 1)
	if err == nil {
		t.Error("Expected error for price exceeding max")
	}
}

// ==================== Safe Addition Tests ====================

func TestSafeAddUint64Normal(t *testing.T) {
	result, err := utils.SafeAddUint64(100, 200)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 300 {
		t.Errorf("Expected 300, got %d", result)
	}
}

func TestSafeAddUint64Zero(t *testing.T) {
	result, err := utils.SafeAddUint64(0, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

func TestSafeAddUint64Overflow(t *testing.T) {
	// MaxUint64 + 1 should overflow
	maxUint64 := ^uint64(0)
	_, err := utils.SafeAddUint64(maxUint64, 1)
	if err == nil {
		t.Error("Expected overflow error")
	}
	if err != utils.ErrPriceOverflow {
		t.Errorf("Expected ErrPriceOverflow, got %v", err)
	}
}

func TestSafeAddUint64NearOverflow(t *testing.T) {
	// MaxUint64 + 0 should not overflow
	maxUint64 := ^uint64(0)
	result, err := utils.SafeAddUint64(maxUint64, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != maxUint64 {
		t.Errorf("Expected %d, got %d", maxUint64, result)
	}
}

func TestSafeAddUint64LargeNumbers(t *testing.T) {
	// Two large numbers that together would overflow
	maxUint64 := ^uint64(0)
	half := maxUint64 / 2
	_, err := utils.SafeAddUint64(half+1, half+1)
	if err == nil {
		t.Error("Expected overflow error for two large numbers")
	}
}

// ==================== Record Length Validation Tests ====================

func TestValidateRecordLengthValid(t *testing.T) {
	err := utils.ValidateRecordLength(100)
	if err != nil {
		t.Errorf("Expected no error for valid record length, got %v", err)
	}
}

func TestValidateRecordLengthZero(t *testing.T) {
	err := utils.ValidateRecordLength(0)
	if err == nil {
		t.Error("Expected error for zero record length")
	}
}

func TestValidateRecordLengthMax(t *testing.T) {
	err := utils.ValidateRecordLength(uint64(utils.MaxRecordSize))
	if err != nil {
		t.Errorf("Expected no error for max record size, got %v", err)
	}
}

func TestValidateRecordLengthTooLarge(t *testing.T) {
	err := utils.ValidateRecordLength(uint64(utils.MaxRecordSize) + 1)
	if err == nil {
		t.Error("Expected error for record exceeding max size")
	}
}

// ==================== Archive File Count Validation Tests ====================

func TestValidateArchiveFileCountValid(t *testing.T) {
	err := utils.ValidateArchiveFileCount(10)
	if err != nil {
		t.Errorf("Expected no error for valid file count, got %v", err)
	}
}

func TestValidateArchiveFileCountZero(t *testing.T) {
	err := utils.ValidateArchiveFileCount(0)
	if err == nil {
		t.Error("Expected error for zero file count")
	}
}

func TestValidateArchiveFileCountMax(t *testing.T) {
	err := utils.ValidateArchiveFileCount(utils.MaxFileCount)
	if err != nil {
		t.Errorf("Expected no error for max file count, got %v", err)
	}
}

func TestValidateArchiveFileCountTooMany(t *testing.T) {
	err := utils.ValidateArchiveFileCount(utils.MaxFileCount + 1)
	if err == nil {
		t.Error("Expected error for exceeding max file count")
	}
}

// ==================== Decompressed Size Validation Tests ====================

func TestValidateDecompressedSizeValid(t *testing.T) {
	err := utils.ValidateDecompressedSize(1024)
	if err != nil {
		t.Errorf("Expected no error for valid decompressed size, got %v", err)
	}
}

func TestValidateDecompressedSizeMax(t *testing.T) {
	err := utils.ValidateDecompressedSize(utils.MaxDecompressedSize)
	if err != nil {
		t.Errorf("Expected no error for max decompressed size, got %v", err)
	}
}

func TestValidateDecompressedSizeTooLarge(t *testing.T) {
	err := utils.ValidateDecompressedSize(utils.MaxDecompressedSize + 1)
	if err == nil {
		t.Error("Expected error for exceeding max decompressed size")
	}
}

// ==================== Offset Validation Tests ====================

func TestValidateOffsetValid(t *testing.T) {
	err := utils.ValidateOffset(50, 100)
	if err != nil {
		t.Errorf("Expected no error for valid offset, got %v", err)
	}
}

func TestValidateOffsetZero(t *testing.T) {
	err := utils.ValidateOffset(0, 100)
	if err != nil {
		t.Errorf("Expected no error for zero offset, got %v", err)
	}
}

func TestValidateOffsetNegative(t *testing.T) {
	err := utils.ValidateOffset(-1, 100)
	if err == nil {
		t.Error("Expected error for negative offset")
	}
}

func TestValidateOffsetAtFileSize(t *testing.T) {
	err := utils.ValidateOffset(100, 100)
	if err == nil {
		t.Error("Expected error for offset at file size")
	}
}

func TestValidateOffsetBeyondFileSize(t *testing.T) {
	err := utils.ValidateOffset(150, 100)
	if err == nil {
		t.Error("Expected error for offset beyond file size")
	}
}

// ==================== Integration-style Validation Tests ====================

func TestValidationConstantsAreSensible(t *testing.T) {
	// Ensure constants are set to reasonable values
	if utils.MaxNameLength < 50 {
		t.Error("MaxNameLength is too small for practical use")
	}
	if utils.MaxNameLength > 10000 {
		t.Error("MaxNameLength is too large, may cause issues")
	}

	if utils.MaxItemsPerCollection < 10 {
		t.Error("MaxItemsPerCollection is too small")
	}

	if utils.MaxRecordSize < 1024 {
		t.Error("MaxRecordSize is too small")
	}

	if utils.MaxDecompressedSize < utils.MaxRecordSize {
		t.Error("MaxDecompressedSize should be >= MaxRecordSize")
	}
}
